import React, { Fragment } from "react";
import { EuiBadge, EuiComment } from "@elastic/eui";
import { DateFromNow } from "@gojek/mlp-ui";
import { Status } from "../../../services/status/Status";
import PulseLoader from "react-spinners/PulseLoader";
import { ExpandableContainer } from "../../../components/expandable_container/ExpandableContainer";
import "./EventsList.scss";

const infoColor = "#abe095";

export const EventsList = ({ events, status }) => {
  const badge = event => {
    const color = event.event_type === "info" ? infoColor : "danger";
    return <EuiBadge color={color}>{event.stage}</EuiBadge>;
  };

  const timelineIcon = event => {
    switch (event.event_type) {
      case "error":
        return "alert";
      default:
        return undefined;
    }
  };

  return (
    <Fragment>
      {events.map((event, idx) => (
        <EuiComment
          key={`event-${idx}`}
          username={badge(event)}
          timelineIcon={timelineIcon(event)}
          type="update"
          timestamp={<DateFromNow date={event.created_at} size={"s"} />}>
          <ExpandableContainer
            maxHeight={43}
            toggleKind="text"
            isScrollable={false}>
            <span className="activityLogEvent">{event.message}</span>
          </ExpandableContainer>
        </EuiComment>
      ))}
      {status === Status.PENDING && (
        <EuiComment
          username={<PulseLoader size={6} color={infoColor} />}
          type="update"
        />
      )}
    </Fragment>
  );
};
