import React, { useEffect, useMemo } from "react";
import {
  EuiPanel,
  EuiTextAlign,
  EuiLoadingChart,
  EuiCallOut,
} from "@elastic/eui";
import { ConfigSection } from "../../components/config_section";
import ScrollToBottom from "react-scroll-to-bottom";
import { useInitiallyLoaded } from "../../hooks/useInitiallyLoaded";
import { Status } from "../../services/status/Status";
import { EventsList } from "./events_list/EventsList";
import "./RouterActivityLogView.scss";
import { useTuringPollingApi } from "../../hooks/useTuringPollingApi";

const POLLING_INTERVAL = 5000;
export const RouterActivityLogView = ({ projectId, routerId, router }) => {
  const [
    {
      data: { events },
      isLoaded,
      error,
    },
    startPollingEvents,
    stopPollingEvents,
    fetchEventsOnce,
  ] = useTuringPollingApi(
    `/projects/${projectId}/routers/${routerId}/events`,
    {},
    [],
    POLLING_INTERVAL
  );
  const hasInitiallyLoaded = useInitiallyLoaded(isLoaded);

  const routerStatus = useMemo(
    () => Status.fromValue(router.status),
    [router.status]
  );

  useEffect(() => {
    fetchEventsOnce();
  }, [fetchEventsOnce]);

  useEffect(() => {
    if (routerStatus === Status.PENDING) {
      startPollingEvents();
    }
    return stopPollingEvents;
  }, [routerStatus, startPollingEvents, stopPollingEvents]);

  return (
    <ConfigSection title="Activity Log">
      <EuiPanel className="euiPanel--activityLog">
        {!hasInitiallyLoaded ? (
          <EuiTextAlign textAlign="center">
            <EuiLoadingChart size="xl" mono />
          </EuiTextAlign>
        ) : error ? (
          <EuiCallOut
            title="Sorry, there was an error"
            color="danger"
            iconType="alert"
          >
            <p>{error.message}</p>
          </EuiCallOut>
        ) : (
          <ScrollToBottom
            className="scrollToBottom--container"
            followButtonClassName="followButton"
          >
            <EventsList events={events} status={routerStatus} />
          </ScrollToBottom>
        )}
      </EuiPanel>
    </ConfigSection>
  );
};
