import NextLink from "next/link";
import React from "react";

/**
 * Link — the navigation primitive other library components route through
 * (PageHeader actions/breadcrumbs, RowActionsMenu href items, ...).
 *
 * Internal hrefs render next/link: client-side transitions, prefetching,
 * and automatic basePath prefixing. External URLs (http(s)://, mailto:,
 * tel:) render a plain <a> — next/link must never handle those.
 */

const EXTERNAL_HREF = /^(?:[a-z][a-z0-9+.-]*:)?\/\//i;

/** True for absolute/external URLs that must bypass client routing. */
function isExternalHref(href: string): boolean {
  return EXTERNAL_HREF.test(href) || href.startsWith("mailto:") || href.startsWith("tel:");
}

export type LinkProps = React.AnchorHTMLAttributes<HTMLAnchorElement> & {
  href: string;
};

export default function Link({ href, children, ...rest }: LinkProps) {
  if (isExternalHref(href)) {
    return (
      <a href={href} {...rest}>
        {children}
      </a>
    );
  }
  return (
    <NextLink href={href} {...rest}>
      {children}
    </NextLink>
  );
}
