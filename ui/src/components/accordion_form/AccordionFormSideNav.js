import React from "react";
import { slugify } from "@gojek/mlp-ui";
import { EuiIcon, EuiSideNav } from "@elastic/eui";
import { Link } from "react-scroll";
import { animatedScrollConfig } from "./scroll";

export const AccordionFormSideNav = ({ name, sections }) => {
  const sideNav = [
    {
      name: name,
      id: 0,
      items: sections.map((section, idx) => ({
        id: idx,
        name: section.title,
        renderItem: () => <AccordionFormSideNavItem section={section} />
      }))
    }
  ];

  return <EuiSideNav items={sideNav} />;
};

const AccordionFormSideNavItem = ({ section }) => (
  <Link
    className="euiSideNavItemButton euiSideNavItemButton--isClickable"
    activeClass="euiSideNavItemButton-isSelected"
    to={`${slugify(section.title)}`}
    spy={true}
    {...animatedScrollConfig}>
    <span className="euiSideNavItemButton__content">
      {section.iconType && (
        <EuiIcon
          type={section.iconType}
          className="euiSideNavItemButton__icon"
          size="m"
        />
      )}
      <span className="euiSideNavItemButton__label">{section.title}</span>
    </span>
  </Link>
);
