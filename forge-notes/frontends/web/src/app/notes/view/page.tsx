"use client";

// yours: scaffolded once, never touched again — forge will not overwrite this file
// Disposable: forge scaffolds a CRUD route for EVERY entity; delete it (and its nav entry) if unused.
// Detail page emitted by `forge scaffold page`, served at
// /notes/view?id=<id>. See `forge skill load frontend/pages` for canonical patterns.
import { userMessage } from "@reliantlabs/forge-web-runtime";
import { Trash2, SquarePen, TriangleAlert } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Suspense, useState } from "react";

import AlertBanner from "@/components/ui/alert_banner";
import KeyValueList from "@/components/ui/key_value_list";
import Modal from "@/components/ui/modal";
import PageHeader from "@/components/ui/page_header";
import SkeletonLoader from "@/components/ui/skeleton_loader";
import { useDeleteNote, useGetNote } from "@/hooks/notes-service-hooks_gen";
import { entityEditHref, useEntityIdParam } from "@/lib/entity-routes";
import { formatValue } from "@/lib/format-utils";

// The id rides in the query string, not a `[id]` path segment, so this page
// is ONE static route that `output: "export"` can build — the id exists only
// in the browser (src/lib/entity-routes.ts). Reading it calls
// useSearchParams, which must sit under a Suspense boundary: the export
// prerenders the fallback, and the browser renders the row.
export default function NoteDetailPage() {
  return (
    <Suspense fallback={<SkeletonLoader variant="list-item" count={5} />}>
      <NoteDetailRoute />
    </Suspense>
  );
}

// The query string is an external boundary — validate, don't cast (the
// `frontend/patterns` skill, "Typed boundaries"). A link that lost its id
// gets a way back, not a crash.
function NoteDetailRoute() {
  const id = useEntityIdParam();
  if (id === undefined) {
    return (
      <div className="px-4 py-8">
        <AlertBanner variant="warning" message="This link doesn't say which note to show." />
        <Link
          href="/notes"
          className="mt-4 inline-flex text-sm text-accent underline hover:text-accent-hover"
        >
          Back to notes
        </Link>
      </div>
    );
  }
  return <NoteDetail id={id} />;
}

function NoteDetail({ id }: { id: string }) {
  const router = useRouter();
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);

  const { data, isLoading, error } = useGetNote({ id });
  // The generated hook already invalidates the note query scope on
  // success — no manual cache invalidation needed here. Failure surfaces
  // via the app-wide error-toast chokepoint (query-client.ts).
  const deleteMutation = useDeleteNote({
    onSuccess: () => router.push("/notes"),
  });

  if (isLoading) return <SkeletonLoader variant="list-item" count={5} />;
  if (error)
    return (
      <div className="px-4 py-8">
        <AlertBanner variant="error" message={userMessage(error)} />
        <Link
          href="/notes"
          className="mt-4 inline-flex text-sm text-accent underline hover:text-accent-hover"
        >
          Back to notes
        </Link>
      </div>
    );

  // Typed access: the generator knows the response wraps the entity in
  // `note` — no `?? data` hedging, no Record<string, unknown> cast.
  const item = data?.note;
  const title = item?.title || String(item?.id ?? "Note");

  // Typed field rows — one entry per renderable proto field, declared by
  // the generator instead of reflected out of Object.entries at runtime.
  const fields = item
    ? [
        {
          label: "Title",
          value: formatValue(item.title),
        },
        {
          label: "Body",
          value: formatValue(item.body),
        },
        {
          label: "Id",
          value: formatValue(item.id),
        },
        {
          label: "Created At",
          value: formatValue(item.createdAt),
        },
        {
          label: "Updated At",
          value: formatValue(item.updatedAt),
        },
      ]
    : [];

  const actions: Array<{
    label: string;
    href?: string;
    variant?: "primary" | "secondary" | "danger";
    icon?: React.ReactNode;
    onClick?: () => void;
  }> = [
    {
      label: "Edit",
      href: entityEditHref("notes", id),
      variant: "secondary" as const,
      icon: <SquarePen aria-hidden className="h-4 w-4" strokeWidth={1.5} />,
    },
    {
      label: "Delete",
      onClick: () => setShowDeleteConfirm(true),
      variant: "danger" as const,
      icon: <Trash2 aria-hidden className="h-4 w-4" strokeWidth={1.5} />,
    },
  ];

  return (
    <div>
      <PageHeader
        title={title}
        breadcrumbs={[{ label: "Notes", href: "/notes" }, { label: title }]}
        actions={actions}
      />

      <KeyValueList groups={[{ fields }]} />
      <Modal
        open={showDeleteConfirm}
        onClose={() => setShowDeleteConfirm(false)}
        title="Delete Note"
        description="Are you sure you want to delete this note? This action cannot be undone."
        footer={
          <>
            <button
              onClick={() => setShowDeleteConfirm(false)}
              className="rounded-lg border border-border bg-surface px-4 py-2 text-sm font-medium text-ink shadow-sm hover:bg-surface-muted"
            >
              Cancel
            </button>
            <button
              onClick={() => deleteMutation.mutate({ id })}
              disabled={deleteMutation.isPending}
              className="rounded-lg bg-danger px-4 py-2 text-sm font-medium text-on-danger shadow-sm hover:bg-danger-hover disabled:opacity-50"
            >
              {deleteMutation.isPending ? "Deleting..." : "Delete"}
            </button>
          </>
        }
      >
        <div className="flex h-10 w-10 items-center justify-center rounded-full bg-danger-surface">
          <TriangleAlert aria-hidden className="h-5 w-5 text-danger" strokeWidth={1.5} />
        </div>
      </Modal>
    </div>
  );
}
