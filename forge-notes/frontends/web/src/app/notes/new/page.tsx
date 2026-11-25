"use client";

// yours: scaffolded once, never touched again — forge will not overwrite this file
// Disposable: forge scaffolds a CRUD route for EVERY entity; delete it (and its nav entry) if unused.
// Create page emitted by `forge scaffold page`. See `forge skill load frontend/pages` for canonical patterns.
import { zodResolver } from "@hookform/resolvers/zod";
import { userMessage } from "@reliantlabs/forge-web-runtime";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { z } from "zod";

import AlertBanner from "@/components/ui/alert_banner";
import Button from "@/components/ui/button";
import Input from "@/components/ui/input";
import Label from "@/components/ui/label";
import PageHeader from "@/components/ui/page_header";
import { useCreateNote } from "@/hooks/notes-service-hooks_gen";
import { entityViewHref } from "@/lib/entity-routes";

const schema = z.object({
  title: z.string(),
  body: z.string(),
});

type FormInput = z.input<typeof schema>;
type FormValues = z.output<typeof schema>;

export default function CreateNotePage() {
  const router = useRouter();

  const mutation = useCreateNote({
    // Land on the record just created — its detail page is where the next
    // step (editing it, adding a related row) starts.
    onSuccess: ({ note: created }) =>
      router.push(created ? entityViewHref("notes", created.id) : "/notes"),
    // The error renders inline below — opt out of the app-wide error
    // toast so the failure isn't announced twice.
    meta: { silenceErrorToast: true },
  });
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<FormInput, unknown, FormValues>({
    resolver: zodResolver(schema),
  });

  function onSubmit(values: FormValues) {
    // Typed end-to-end: every form value that is not already the wire type
    // is converted BEFORE mutate() — date strings to proto Timestamps,
    // 64-bit ints to bigints, base64 text to bytes, comma-separated text
    // to arrays — so the request matches the schema instead of being cast
    // into shape.
    mutation.mutate({
      ...values,
    });
  }

  return (
    <div className="mx-auto max-w-xl px-4 py-8">
      <PageHeader
        title="Create Note"
        breadcrumbs={[{ label: "Notes", href: "/notes" }, { label: "Create" }]}
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
          <Button href="/notes" variant="outline" size="lg">
            Cancel
          </Button>
          <Button
            type="submit"
            size="lg"
            loading={mutation.isPending}
            disabled={mutation.isPending}
          >
            {mutation.isPending ? "Creating…" : "Create Note"}
          </Button>
        </div>
      </form>
    </div>
  );
}
