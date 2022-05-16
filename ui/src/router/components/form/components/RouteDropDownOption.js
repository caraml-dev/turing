import {
  EuiBadge,
  EuiFlexGroup,
  EuiFlexItem,
  EuiTextColor,
  EuiText,
  EuiToolTip,
} from "@elastic/eui";

export const RouteDropDownOption = ({
  id,
  endpoint,
  isDisabled,
  disabledOptionTooltip,
}) => {
  const option = (
    <EuiFlexGroup direction="row" gutterSize="s">
      <EuiFlexItem grow={false}>
        <EuiBadge color={isDisabled ? "hollow" : "default"}>{id}</EuiBadge>
      </EuiFlexItem>
      <EuiFlexItem className="eui-textTruncate">
        <EuiTextColor color="subdued">
          <EuiText size="s" className="eui-textTruncate">
            {endpoint}
          </EuiText>
        </EuiTextColor>
      </EuiFlexItem>
    </EuiFlexGroup>
  );

  return isDisabled ? (
    <EuiToolTip content={disabledOptionTooltip}>{option}</EuiToolTip>
  ) : (
    option
  );
};
