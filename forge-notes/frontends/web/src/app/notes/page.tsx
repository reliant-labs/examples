"use client";

// yours: scaffolded once, never touched again — forge will not overwrite this file
// Disposable: forge scaffolds a CRUD route for EVERY entity; delete it (and its nav entry) if unused.
// Notes list page (`forge scaffold page`): <Resource> owns loading/error/empty states,
// the filter box and cursor pagination, all SERVER-SIDE via the List RPC's filter fields — never
// filter one fetched page on the client. Depth: `forge skill load frontend/pages`.

import { Resource, type ResourceColumn, type ResourceSort } from "@reliantlabs/forge-web-runtime";
import { keepPreviousData } from "@tanstack/react-query";
import { Plus } from "lucide-react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { Suspense, useCallback, useState } from "react";
import { z } from "zod";

import Button from "@/components/ui/button";
import SkeletonLoader from "@/components/ui/skeleton_loader";
import { useListNotes } from "@/hooks/notes-service-hooks_gen";
import { useQueryResource } from "@/hooks/use-query-resource";
import { entityViewHref } from "@/lib/entity-routes";
import { formatValue } from "@/lib/format-utils";
import { useTypedSearchParams } from "@/lib/search-schemas";

import type { Note } from "@/gen/services/notes/v1/notes_pb";

// Rows per page requested from the server. Paginate the result with the
// <Resource> footer's Prev/Next (they walk the next_page_token cursor).
// Bump this to taste.
const PAGE_SIZE = 25;

// ---- Typed columns ----
// One entry per renderable proto field, type-checked against Note.
// The id and the created/updated timestamps are deliberately NOT here — the
// primary cell already links to the detail page, which is where they belong.
// Add one back if this list needs it; this file is yours.
// A column with a sortKey sorts SERVER-SIDE: the key is the List request's
// order_by (a column name). A sorted list is one page — the server's keyset
// cursor pages only in primary-key order, so it returns no next-page token.
const columns: ResourceColumn<Note>[] = [
  {
    header: "Title",
    sortKey: "title",
    cell: (item) => (
      <Link
        href={entityViewHref("notes", item.id)}
        className="font-medium text-ink underline-offset-4 hover:underline focus-visible:rounded-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent"
      >
        {formatValue(item.title)}
      </Link>
    ),
  },
  {
    header: "Body",
    sortKey: "body",
    cell: (item) => formatValue(item.body),
  },
];

// The URL is the source of truth for this page's filters and sort — the
// "schema at the URL boundary" convention from the `frontend/state` skill,
// driven by useTypedSearchParams. A dashboard tile linking to
// `/notes?…` therefore lands on a FILTERED list; the Back button undoes a
// filter change; and a filtered view is a link somebody can paste into a
// ticket. Holding these in useState instead makes every such link dead — it
// renders the unfiltered list, with no error anywhere to say so.
const searchParamsSchema = z.object({
  q: z.string().optional(),
  sort: z.string().optional(),
  desc: z.string().optional(),
});

// The sort keys this page accepts from `?sort=`. Anything else is ignored
// rather than sent: the server refuses an order_by it does not know.
const sortKeys = ["title", "body"] as const;

function NotesPageContent() {
  const router = useRouter();
  const pathname = usePathname();
  const params = useTypedSearchParams(searchParamsSchema);

  // One writer for every filter. It reads the LIVE query string instead of
  // closing over the parsed params, which keeps the callback stable — so the
  // debounced effects below fire on a real change, not on every navigation.
  const setFilters = useCallback(
    (next: Record<string, string | undefined>) => {
      const sp = new URLSearchParams(window.location.search);
      for (const [key, value] of Object.entries(next)) {
        if (value === undefined || value === "") sp.delete(key);
        else sp.set(key, value);
      }
      const qs = sp.toString();
      router.replace(qs ? `${pathname}?${qs}` : pathname, { scroll: false });
    },
    [pathname, router],
  );

  // Cursor stack: each entry is the page_token that opened that page (the
  // last entry is the current page; empty = first page). The cursor stays in
  // component state — it is an opaque server cursor, not a view worth
  // linking to — and any filter or sort change resets the stack to page one.
  const [cursor, setCursor] = useState<string[]>([]);
  const search = params.q ?? "";
  const sortKey = sortKeys.find((key) => key === params.sort);
  const sort: ResourceSort | null = sortKey
    ? { key: sortKey, descending: params.desc === "true" }
    : null;

  // Server-side query: filters, sort and cursor go to the backend; nothing is
  // filtered, sorted or capped in the browser. keepPreviousData holds the
  // prior page while the next loads (no skeleton flash).
  const query = useListNotes(
    {
      pageSize: PAGE_SIZE,
      pageToken: cursor.at(-1),
      orderBy: sort?.key ?? "",
      descending: sort?.descending ?? false,
      search: search || undefined,
    },
    { placeholderData: keepPreviousData },
  );
  const resource = useQueryResource(query);

  const nextToken = resource.status === "success" ? resource.data.nextPageToken : "";
  const total = resource.status === "success" ? Number(resource.data.totalCount) : undefined;

  return (
    <Resource<Note>
      title="Notes"
      status={resource.status}
      data={resource.status === "success" ? resource.data.notes : undefined}
      error={resource.status === "error" ? resource.error : undefined}
      onRetry={resource.status === "error" ? resource.refetch : undefined}
      isFetching={query.isFetching}
      columns={columns}
      rowKey={(item) => String(item.id)}
      onRowClick={(item) => router.push(entityViewHref("notes", item.id))}
      rowsHaveLinks
      emptyTitle="No notes found"
      emptyMessage="No records match the current view."
      filter={search}
      onFilterChange={(value) => {
        setFilters({ q: value });
        setCursor([]);
      }}
      filterPlaceholder="Search notes…"
      sort={sort}
      onSortChange={(next) => {
        setFilters({ sort: next?.key, desc: next?.descending ? "true" : undefined });
        setCursor([]);
      }}
      onNextPage={() => setCursor((c) => [...c, nextToken])}
      onPrevPage={() => setCursor((c) => c.slice(0, -1))}
      hasNextPage={Boolean(nextToken)}
      hasPrevPage={cursor.length > 0}
      totalCount={total}
      pageStart={cursor.length * PAGE_SIZE}
      actions={
        <Button href="/notes/new" size="lg" leftIcon={<Plus aria-hidden strokeWidth={2} />}>
          New Note
        </Button>
      }
    />
  );
}

// useSearchParams() reads a value that only exists on the client, so Next
// refuses to prerender a route that calls it outside a <Suspense> boundary —
// `next build` fails the page otherwise. The prop-free wrapper below is the
// route's actual export; the content component above is where the page lives.
export default function NotesPage() {
  return (
    <Suspense fallback={<SkeletonLoader variant="table-row" count={5} />}>
      <NotesPageContent />
    </Suspense>
  );
}
