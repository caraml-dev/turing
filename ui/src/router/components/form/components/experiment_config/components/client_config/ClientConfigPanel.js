import { Panel } from "../../../Panel";
import React, { useCallback, useContext, useEffect, useMemo } from "react";
import { EuiFieldText, EuiForm, EuiFormRow, EuiSpacer } from "@elastic/eui";
import { EuiComboBoxSelect } from "../../../../../../../components/form/combo_box/EuiComboBoxSelect";
import ExperimentContext from "../../providers/context";
import { FormLabelWithToolTip } from "../../../../../../../components/form/label_with_tooltip/FormLabelWithToolTip";
import sortBy from "lodash/sortBy";

export const ClientConfigPanel = ({ client, onChangeHandler, errors = {} }) => {
  const { clients, isLoading, isLoaded, setClientsValidated } = useContext(
    ExperimentContext
  );

  const clientsOptions = useMemo(() => {
    return sortBy(clients, "username").map(c => ({
      icon: "user",
      label: c.username
    }));
  }, [clients]);

  // Clear the current client if the new client list does not include it
  useEffect(() => {
    if (isLoaded("clients")) {
      if (
        !!client.username &&
        !clients.find(c => c.username === client.username)
      ) {
        onChangeHandler({});
      }
      setClientsValidated(true);
    }
  }, [
    client.username,
    clients,
    isLoaded,
    setClientsValidated,
    onChangeHandler
  ]);

  // Define onchange handlers
  const onClientUsernameChange = useCallback(
    clientUsername => {
      if (clientUsername !== client.username) {
        const clientInfo =
          clients.find(c => c.username === clientUsername) || {};
        onChangeHandler({ ...clientInfo });
      }
    },
    [client, clients, onChangeHandler]
  );

  const onPasskeyChange = useCallback(
    event => {
      if (!client.passkey_set) {
        client.passkey = event.target.value;
      } else {
        client.passkey_set = false;
        client.passkey = event.nativeEvent.data || "";
      }
      onChangeHandler(client);
    },
    [client, onChangeHandler]
  );

  return (
    <Panel title="Client Credentials">
      <EuiSpacer size="m" />
      <EuiForm>
        <EuiFormRow
          fullWidth
          label={
            <FormLabelWithToolTip
              label="Client ID *"
              content="Select your Client ID from the provided list"
            />
          }
          isInvalid={!!errors.id}
          error={errors.id}
          display="row">
          <EuiComboBoxSelect
            fullWidth
            isLoading={isLoading("clients")}
            placeholder="Select"
            value={client.username || ""}
            options={clientsOptions}
            onChange={onClientUsernameChange}
            isInvalid={!!errors.username}
          />
        </EuiFormRow>

        <EuiSpacer size="m" />

        <EuiFormRow
          fullWidth
          label="Passkey *"
          isInvalid={!!errors.passkey}
          error={errors.passkey}>
          <EuiFieldText
            fullWidth
            placeholder="passkey"
            value={client.passkey || ""}
            onChange={onPasskeyChange}
            isInvalid={!!errors.passkey}
          />
        </EuiFormRow>
      </EuiForm>
    </Panel>
  );
};
