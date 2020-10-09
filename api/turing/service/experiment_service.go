package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"

	"github.com/gojek/turing/api/turing/config"
	"github.com/gojek/turing/api/turing/models"
	"github.com/gojek/turing/engines/experiment/common"
	"github.com/gojek/turing/engines/experiment/manager"
)

const (
	expCacheExpirySeconds  = 600
	expCacheCleanUpSeconds = 900
)

// ExperimentsService provides functionality to work with experiment engines
// supported by Turing
type ExperimentsService interface {
	// ListEngines returns a list of the experiment engines available
	ListEngines() []manager.Engine
	// ListClients returns a list of the clients registered on the given experiment engine
	ListClients(engine string) ([]manager.Client, error)
	// ListExperiments returns a list of the experiments registered on the given experiment engine,
	// and for the given clientID if supplied
	ListExperiments(engine string, clientID string) ([]manager.Experiment, error)
	// ListVariables returns a list of the variables registered on the given experiment engine,
	// for the given clientID and/or experiments
	ListVariables(engine string, clientID string, experimentIDs []string) (manager.Variables, error)
	// ValidateExperimentConfig validates the given experiment config for completeness
	ValidateExperimentConfig(engine string, cfg manager.TuringExperimentConfig) error
	// GetExperimentRunnerConfig converts the given generic manager.TuringExperimentConfig into the
	// format compatible with the ExperimentRunner
	GetExperimentRunnerConfig(engine string, cfg *manager.TuringExperimentConfig) (json.RawMessage, error)
}

type experimentsService struct {
	// map of engine name -> Experiment Manager
	experimentManagers map[models.ExperimentEngineType]manager.ExperimentManager
	cache              *cache.Cache
}

// Experiment represents an experiment in Turing. The experiment info can come from different expriment engines
// such as Litmus or XP but they will be represented with the same schema in this experiment object. For instance
// in Litmus, the variation of the same experiment is called "variants" whereas in XP it is called "treament", but in
// Turing the term "treatment" in an experiment is consistently used.
type Experiment struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	ClientName string   `json:"client_name"`
	UnitType   string   `json:"unit_type"`
	Treatments []string `json:"treatments"` // List of treatment names (i.e variations) in the experiment.
}

// NewExperimentsService creates a new experiment service from the given config
func NewExperimentsService(litmusCfg *config.LitmusConfig, xpCfg *config.XPConfig) (ExperimentsService, error) {
	experimentManagers := make(map[models.ExperimentEngineType]manager.ExperimentManager)

	// Initialize the experimentsService with cache
	svc := &experimentsService{
		experimentManagers: experimentManagers,
		cache:              cache.New(expCacheExpirySeconds*time.Second, expCacheCleanUpSeconds*time.Second),
	}

	// Populate the cache with the Clients / Experiments info
	for expEngine, expManager := range svc.experimentManagers {
		if expManager.GetEngineInfo().ClientSelectionEnabled {
			_, err := svc.ListClients(string(expEngine))
			if err != nil {
				return nil, err
			}
		} else if expManager.GetEngineInfo().ExperimentSelectionEnabled {
			_, err := svc.ListExperiments(string(expEngine), "")
			if err != nil {
				return nil, err
			}
		}
	}

	return svc, nil
}

func (es *experimentsService) ListEngines() []manager.Engine {
	engines := []manager.Engine{}

	for _, expManager := range es.experimentManagers {
		engines = append(engines, expManager.GetEngineInfo())
	}
	return engines
}

func (es *experimentsService) ListClients(engine string) ([]manager.Client, error) {
	return es.listClientsWithCache(engine)
}

func (es *experimentsService) ListExperiments(
	engine string,
	clientID string,
) ([]manager.Experiment, error) {
	// Get client, if the clientID has been supplied
	var client *manager.Client
	var err error
	if clientID != "" {
		client, err = es.getClientWithCache(engine, clientID)
		if err != nil {
			return []manager.Experiment{}, err
		}
	}

	return es.listExperimentsWithCache(engine, client)
}

func (es *experimentsService) ListVariables(
	engine string,
	clientID string,
	experimentIDs []string,
) (manager.Variables, error) {
	var variables manager.Variables
	var err error

	// Get client and its variables, if the clientID has been supplied
	var client *manager.Client
	clientVariables := []manager.Variable{}
	if clientID != "" {
		client, err = es.getClientWithCache(engine, clientID)
		if err != nil {
			return variables, err
		}
		clientVariables, err = es.listClientVariablesWithCache(engine, client)
		if err != nil {
			return variables, err
		}
	}

	// Get experiments and their variables, if experiment ids have been supplied
	var experiments []manager.Experiment
	experimentVariables := make(map[string][]manager.Variable)
	if len(experimentIDs) > 0 {
		experiments, err = es.getExperimentsWithCache(engine, client, experimentIDs)
		if err != nil {
			return variables, err
		}
		experimentVariables, err = es.listExperimentVariablesWithCache(engine, experiments)
		if err != nil {
			return variables, err
		}
	}

	// Reconcile variables for configuration
	variableConfigs := reconcileVariables(clientVariables, experimentVariables)

	// Return variables info
	return manager.Variables{
		ClientVariables:     clientVariables,
		ExperimentVariables: experimentVariables,
		Config:              variableConfigs,
	}, nil
}

func (es *experimentsService) ValidateExperimentConfig(engine string, cfg manager.TuringExperimentConfig) error {
	// Get experiment manager
	expManager, err := es.getExperimentManager(engine)
	if err != nil {
		return err
	}

	return expManager.ValidateExperimentConfig(expManager.GetEngineInfo(), cfg)
}

func (es *experimentsService) GetExperimentRunnerConfig(
	engine string,
	cfg *manager.TuringExperimentConfig,
) (json.RawMessage, error) {
	// Get experiment manager
	expManager, err := es.getExperimentManager(engine)
	if err != nil {
		return json.RawMessage{}, err
	}

	return expManager.GetExperimentRunnerConfig(*cfg)
}

func (es *experimentsService) getExperimentManager(
	engine string,
) (manager.ExperimentManager, error) {
	expManager, ok := es.experimentManagers[models.ExperimentEngineType(engine)]
	if !ok {
		return nil, fmt.Errorf("Unknown experiment engine %s", engine)
	}
	return expManager, nil
}

func (es *experimentsService) listClientsWithCache(engine string) ([]manager.Client, error) {
	// Get experiment manager
	expManager, err := es.getExperimentManager(engine)
	if err != nil {
		return []manager.Client{}, err
	}

	cacheKey := fmt.Sprintf("engine:%s:clients", engine)
	cacheEnabled := expManager.IsCacheEnabled()

	if cacheEnabled {
		// Attempt to retrieve info from cache
		cachedValue, found := es.cache.Get(cacheKey)
		// Found in cache - cast the data and return
		if found {
			clients, ok := cachedValue.([]manager.Client)
			if !ok {
				return []manager.Client{},
					fmt.Errorf("Malformed clients info found in the cache for engine %s", engine)
			}
			return clients, nil
		}
	}

	// Cache disabled / not found in cache - invoke API
	clients, err := expManager.ListClients()
	if err != nil {
		return []manager.Client{}, err
	}
	if cacheEnabled {
		es.cache.Set(cacheKey, clients, cache.DefaultExpiration)
	}

	return clients, nil
}

func (es *experimentsService) listExperimentsWithCache(
	engine string,
	client *manager.Client,
) ([]manager.Experiment, error) {
	// Get experiment manager
	expManager, err := es.getExperimentManager(engine)
	if err != nil {
		return []manager.Experiment{}, err
	}

	var cacheKey string
	var listExperimentsMethod func() ([]manager.Experiment, error)
	cacheEnabled := expManager.IsCacheEnabled()

	// Set cache key and API method to be called, based on client info passed in
	if client != nil {
		cacheKey = fmt.Sprintf("engine:%s:clients:%s:experiments", engine, client.ID)
		listExperimentsMethod = func() ([]manager.Experiment, error) {
			return expManager.ListExperimentsForClient(*client)
		}
	} else {
		cacheKey = fmt.Sprintf("engine:%s:experiments", engine)
		listExperimentsMethod = func() ([]manager.Experiment, error) {
			return expManager.ListExperiments()
		}
	}

	if cacheEnabled {
		// Attempt to retrieve info from cache
		cachedValue, found := es.cache.Get(cacheKey)
		// Found in cache - cast the data and return
		if found {
			experiments, ok := cachedValue.([]manager.Experiment)
			if !ok {
				return []manager.Experiment{},
					errors.New("Malformed experiments info found in the cache")
			}
			return experiments, nil
		}
	}

	// Cache disabled / not found in cache - invoke Experiment Manager API
	experiments, err := listExperimentsMethod()
	if err != nil {
		return []manager.Experiment{}, err
	}
	if cacheEnabled {
		es.cache.Set(cacheKey, experiments, cache.DefaultExpiration)
	}

	return experiments, nil
}

func (es *experimentsService) listClientVariablesWithCache(
	engine string,
	client *manager.Client,
) ([]manager.Variable, error) {
	// Get experiment manager
	expManager, err := es.getExperimentManager(engine)
	if err != nil {
		return []manager.Variable{}, err
	}

	cacheKey := fmt.Sprintf("engine:%s:clients:%s:variables", engine, client.ID)
	cacheEnabled := expManager.IsCacheEnabled()

	if cacheEnabled {
		// Attempt to retrieve info from cache
		cachedValue, found := es.cache.Get(cacheKey)
		// Found in cache - cast the data and return
		if found {
			variables, ok := cachedValue.([]manager.Variable)
			if !ok {
				return []manager.Variable{},
					fmt.Errorf("Malformed variables info found in the cache for client %s", client.ID)
			}
			return variables, nil
		}
	}

	// Cache disabled / not found in cache - invoke API
	variables, err := expManager.ListVariablesForClient(*client)
	if err != nil {
		return []manager.Variable{}, err
	}
	if cacheEnabled {
		es.cache.Set(cacheKey, variables, cache.DefaultExpiration)
	}

	return variables, nil
}

func (es *experimentsService) listExperimentVariablesWithCache(
	engine string,
	experiments []manager.Experiment,
) (map[string][]manager.Variable, error) {
	// Get experiment manager
	expManager, err := es.getExperimentManager(engine)
	if err != nil {
		return map[string][]manager.Variable{}, err
	}

	cacheEnabled := expManager.IsCacheEnabled()
	// Store variables for each experiment (experiment_id -> variables map)
	expVariables := map[string][]manager.Variable{}
	// Save the experiments whose variables need to be queried
	filteredExperiments := []manager.Experiment{}

	if cacheEnabled {
		// For each experiment, attempt to retrieve variables from cache
		for _, exp := range experiments {
			cacheKey := fmt.Sprintf("engine:%s:experiments:%s:variables", engine, exp.ID)
			cachedValue, found := es.cache.Get(cacheKey)

			if found {
				variables, ok := cachedValue.([]manager.Variable)
				if !ok {
					return expVariables,
						fmt.Errorf("Malformed variables info found in the cache for experiment %s", exp.ID)
				}
				expVariables[exp.ID] = variables
			} else {
				// If not exists, add experiment to filteredExperiments, for querying from API
				filteredExperiments = append(filteredExperiments, exp)
			}
		}
	} else {
		// Cache disabled, query for all experiments
		filteredExperiments = experiments
	}

	// Get variables for filteredExperiments which could be an empty list if we got everything
	// from the cache
	filteredExpVariables, err := expManager.ListVariablesForExperiments(filteredExperiments)
	if err != nil {
		return expVariables, err
	}
	// Merge filteredExpVariables and expVariables, saving to cache, and return
	for id, vars := range filteredExpVariables {
		expVariables[id] = vars
		if cacheEnabled {
			cacheKey := fmt.Sprintf("engine:%s:experiments:%s:variables", engine, id)
			es.cache.Set(cacheKey, vars, cache.DefaultExpiration)
		}
	}
	return expVariables, nil
}

func (es *experimentsService) getClientWithCache(
	engine string,
	clientID string,
) (*manager.Client, error) {
	// Get all clients for the engine
	clients, err := es.listClientsWithCache(engine)
	if err != nil {
		return nil, err
	}

	// Filter clients by ID
	for _, c := range clients {
		if c.ID == clientID {
			return &c, nil
		}
	}

	return nil, fmt.Errorf("Could not find client with ID %s", clientID)
}

func (es *experimentsService) getExperimentsWithCache(
	engine string,
	client *manager.Client,
	experimentIDs []string,
) ([]manager.Experiment, error) {
	// Get all experiments (for the client, if supplied)
	experiments, err := es.listExperimentsWithCache(engine, client)
	if err != nil {
		return nil, err
	}

	// Filter the experiments whose ids are given in experimentIDs
	expIDMap := make(map[string]manager.Experiment)
	for _, e := range experiments {
		expIDMap[e.ID] = e
	}
	var filteredExperiments []manager.Experiment
	for _, eID := range experimentIDs {
		if e, found := expIDMap[eID]; found {
			filteredExperiments = append(filteredExperiments, e)
		} else {
			return []manager.Experiment{}, fmt.Errorf("Could not find experiment %s", eID)
		}
	}
	return filteredExperiments, nil
}

func reconcileVariables(
	clientVariables []manager.Variable,
	experimentVariables map[string][]manager.Variable,
) []manager.VariableConfig {
	varCfgMap := map[string]manager.VariableConfig{}
	reconcileFunc := func(item manager.Variable) {
		// Assume that variables are case-sensitive and can be uniquely identified by name
		if cfg, found := varCfgMap[item.Name]; found {
			// If a variable is required for any of the experiments / client, it should be required
			// for the overall Turing experiment
			cfg.Required = cfg.Required || item.Required
			varCfgMap[item.Name] = cfg
		} else {
			varCfgMap[item.Name] = manager.VariableConfig{
				Name:     item.Name,
				Required: item.Required,
				// Set header field source by default
				FieldSource: common.HeaderFieldSource,
			}
		}
	}

	// Process client variables and experiment variables
	for _, v := range clientVariables {
		reconcileFunc(v)
	}
	for _, vars := range experimentVariables {
		for _, v := range vars {
			reconcileFunc(v)
		}
	}

	// Convert map to list and return
	variableConfigs := []manager.VariableConfig{}
	for _, item := range varCfgMap {
		variableConfigs = append(variableConfigs, item)
	}
	return variableConfigs
}
