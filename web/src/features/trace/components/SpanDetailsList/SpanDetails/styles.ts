export const styles = {
  accordion: {
    borderRadius: "8px",
    marginBottom: "8px",
  },
  accordionSummary: {
    flexDirection: "row-reverse",
    height: "60px",
    padding: "8px 8px 8px 25px",
    borderRadius: "8px",
    backgroundColor: "#1B1C21",
    "& .MuiAccordionSummary-expandIconWrapper.Mui-expanded": {
      transform: "rotate(90deg)",
    },
  },
  expandedAccordion: {
    backgroundColor: "#2B2D32",
    borderRadius: "8px 8px 0 0",
  },
  expandArrowIcon: {
    color: "#FFFFFF",
    fontSize: "15px",
  },
  spanFlowIconsContainer: {
    flexDirection: "row",
    alignItems: "center",
    marginRight: "17px",
  },
  spanSourceIcon: {
    marginLeft: "22px",
    width: "22px",
    height: "22px",
  },
  spanFlowArrowIcon: {
    fontSize: "16px",
    margin: "0 8px",
    color: "#96979E",
  },
  spanDestIcon: {
    width: "22px",
    height: "22px",
  },
  spanName: {
    fontWeight: "700",
    fontSize: "14px",
    lineHeight: "20px",
    letterSpacing: "0.15px",
    marginBottom: "8px",
  },
  spanTimes: {
    fontSize: "12px",
    fontWeight: "500",
    letterSpacing: "0.4px",
    color: "#B6B7BE",
  },
  spanTimesDivider: {
    margin: "0 4px",
  },
  accordionDetails: {
    padding: "0",
  },
};