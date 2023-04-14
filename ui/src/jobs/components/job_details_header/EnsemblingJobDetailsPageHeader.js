import React, { useContext } from "react";
import { PageSecondaryHeader } from "../../../router/components/page_header/PageSecondaryHeader";
import { HorizontalDescriptionList } from "../../../components/horizontal_description_list/HorizontalDescriptionList";
import EnsemblersContext from "../../../providers/ensemblers/context";
import { EuiBadge, EuiButtonEmpty, EuiIcon, EuiText } from "@elastic/eui";
import { DateFromNow } from "@caraml-dev/ui-lib";
import { useNavigate } from "react-router-dom";

export const EnsemblingJobDetailsPageHeader = ({ job }) => {
  const navigate = useNavigate();
  const { ensemblers, isLoaded: ensemblersLoaded } =
    useContext(EnsemblersContext);

  const headerItems = [
    {
      title: "Ensembler",
      description: (
        <EuiButtonEmpty
          size="s"
          isLoading={!ensemblersLoaded}
          onClick={() =>
            navigate(`../?ensembler_id=${job.ensembler_id}`)
          }>
          {!!ensemblers[job.ensembler_id] ? (
            <EuiText size="s">
              <EuiIcon type={"aggregate"} size="s" />{" "}
              {ensemblers[job.ensembler_id].name}
            </EuiText>
          ) : (
            <span>Loading&hellip;</span>
          )}
        </EuiButtonEmpty>
      ),
    },
    {
      title: "Environment",
      description: <EuiBadge color="default">{job.environment_name}</EuiBadge>,
      flexProps: {
        grow: 1,
        style: {
          minWidth: "100px",
        },
      },
    },
    {
      title: "Created At",
      description: <DateFromNow date={job.created_at} size="s" />,
      flexProps: {
        grow: 1,
        style: {
          minWidth: "100px",
        },
      },
    },
    {
      title: "Updated At",
      description: <DateFromNow date={job.updated_at} size="s" />,
      flexProps: {
        grow: 1,
        style: {
          minWidth: "100px",
        },
      },
    },
  ];

  return (
    <PageSecondaryHeader>
      <HorizontalDescriptionList listItems={headerItems} />
    </PageSecondaryHeader>
  );
};
