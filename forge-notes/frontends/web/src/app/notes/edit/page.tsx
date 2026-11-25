"use client";

// yours: scaffolded once, never touched again — forge will not overwrite this file
// Disposable: forge scaffolds a CRUD route for EVERY entity; delete it (and its nav entry) if unused.
// Edit page emitted by `forge scaffold page`, served at
// /notes/edit?id=<id>. See `forge skill load frontend/pages` for canonical patterns.
import { zodResolver } from "@hookform/resolvers/zod";
import { userMessage } from "@reliantlabs/forge-web-runtime";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Suspense } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import AlertBanner from "@/components/ui/alert_banner";
import Button from "@/components/ui/button";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import PageHeader from "@/components/ui/page_header";
import SkeletonLoader from "@/components/ui/skeleton_loader";
import { useGetNote, useUpdateNote } from "@/hooks/notes-service-hooks_gen";
import { entityViewHref, useEntityIdParam } from "@/lib/entity-routes";

const schema = z.object({
  title: z.string(),
  body: z.string(),
});

type FormInput = z.input<typeof schema>;
type FormValues = z.output<typeof schema>;

// The id rides in the query string, not a `[id]` path segment, so this page
// is ONE static route that `output: "export"` can build — the id exists only
// in the browser (src/lib/entity-routes.ts). Reading it calls
// useSearchParams, which must sit under a Suspense boundary: the export
// prerenders the fallback, and the browser renders the form.
export default function EditNotePage() {
  return (
    <Suspense
      fallback={
        <div className="mx-auto max-w-xl px-4 py-8">
          <SkeletonLoader variant="form-field" count={4} />
        </div>
      }
    >
      <EditNoteRoute />
    </Suspense>
  );
}

// The query string is an external boundary — validate, don't cast. A link
// that lost its id gets a way back, not a crash.
function EditNoteRoute() {
  const id = useEntityIdParam();
  if (id === undefined) {
    return (
      <div className="mx-auto max-w-xl px-4 py-8">
        <AlertBanner variant="warning" message="This link doesn't say which note to edit." />
        <Link
          href="/notes"
          className="mt-4 inline-flex text-sm text-accent underline hover:text-accent-hover"
        >
          Back to notes
        </Link>
      </div>
    );
  }
  return <EditNoteForm id={id} />;
}

function EditNoteForm({ id }: { id: string }) {
  const router = useRouter();

  const { data, isLoading, error: fetchError } = useGetNote({ id });
  // Typed access: the response wraps the entity in `note` —
  // no `?? data` hedging, no Record<string, unknown> cast.
  const item = data?.note;

  const mutation = useUpdateNote({
    onSuccess: () => router.push(entityViewHref("notes", id)),
    // The error renders inline below — opt out of the app-wide error
    // toast so the failure isn't announced twice.
    meta: { silenceErrorToast: true },
  });
  const {
    register,
    handleSubmit,
    formState: { errors, isDirty },
  } = useForm<FormInput, unknown, FormValues>({
    resolver: zodResolver(schema),
    values: item
      ? {
          title: String(item.title ?? ""),
          body: String(item.body ?? ""),
        }
      : undefined,
  });

  function onSubmit(values: FormValues) {
    // Typed end-to-end: every form value that is not already the wire type
    // is converted BEFORE mutate() — date strings to proto Timestamps,
    // 64-bit ints to bigints, base64 text to bytes, comma-separated text
    // to arrays — so the request matches the schema instead of being cast
    // into shape.
    // AIP-134: the entity rides under `note` and the
    // mask names exactly the fields this form edits — the server writes
    // those columns and nothing else. Without the mask, columns the form
    // doesn't carry would be replaced by the request's zero values.
    mutation.mutate({
      note: {
        id: id,
        ...values,
      },
      updateMask: { paths: ["title", "body"] },
    });
  }

  // Loading is its own state — bail explicitly. Do not derive booleans from
  // `!isLoading && X` and act on them; that gate is false during the load
  // window AND when the load returns a falsy value, which races and wipes
  // real state. See `frontend/state` ("Loading is not the negative of success").
  //
  // For type-checker-enforced exhaustiveness, see `useQueryResource` in
  // `src/hooks/use-query-resource.ts` — it returns a discriminated union
  // `{ status: 'loading' } | { status: 'error', error } | { status: 'success', data }`
  // that forces a switch over all three states.
  if (isLoading)
    return (
      <div className="mx-auto max-w-xl px-4 py-8">
        <SkeletonLoader variant="form-field" count={4} />
      </div>
    );

  if (fetchError) {
    return (
      <div className="mx-auto max-w-xl px-4 py-8">
        <AlertBanner variant="error" message={userMessage(fetchError)} />
        <Link
          href="/notes"
          className="mt-4 inline-flex text-sm text-accent underline hover:text-accent-hover"
        >
          Back to notes
        </Link>
      </div>
    );
  }

  const title = item?.title || String(item?.id ?? id);

  return (
    <div className="mx-auto max-w-xl px-4 py-8">
      <PageHeader
        title="Edit Note"
        breadcrumbs={[
          { label: "Notes", href: "/notes" },
          { label: title, href: entityViewHref("notes", id) },
          { label: "Edit" },
        ]}
      />

      {mutation.error && (
        <div className="mb-6">
          <AlertBanner
            variant="error"
            message={userMessage(mutation.error)}
            dismissible
            onDismiss={() => mutation.reset()}
          />
        </div>
      )}

      <form
        onSubmit={handleSubmit(onSubmit)}
        className="space-y-6 rounded-lg border border-border bg-surface p-6 shadow-sm"
      >
        <div>
          <Label htmlFor="title">Title</Label>
          <Input
            type="text"
            id="title"
            {...register("title")}
            invalid={Boolean(errors.title)}
            aria-invalid={errors.title ? true : undefined}
            aria-describedby={errors.title ? "title-error" : undefined}
          />
          {errors.title && (
            <p id="title-error" className="mt-1 text-sm text-danger">
              {errors.title.message}
            </p>
          )}
        </div>
        <div>
          <Label htmlFor="body">Body</Label>
          <Input
            type="text"
            id="body"
            {...register("body")}
            invalid={Boolean(errors.body)}
            aria-invalid={errors.body ? true : undefined}
            aria-describedby={errors.body ? "body-error" : undefined}
          />
          {errors.body && (
            <p id="body-error" className="mt-1 text-sm text-danger">
              {errors.body.message}
            </p>
          )}
        </div>

        <div className="flex items-center justify-end gap-3 border-t border-border pt-6">
          <Button href={entityViewHref("notes", id)} variant="outline" size="lg">
            Cancel
          </Button>
          <Button
            type="submit"
            size="lg"
            loading={mutation.isPending}
            disabled={mutation.isPending || !isDirty}
          >
            {mutation.isPending ? "Saving…" : "Save Changes"}
          </Button>
        </div>
      </form>
    </div>
  );
}
