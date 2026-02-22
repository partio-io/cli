# Render Plan as Markdown with Copy & Download Buttons

## Context

The Plan tab on the checkpoint detail page currently renders plan content as plain text (`whitespace-pre-wrap`). Since plans are stored as `.md` files, they should be rendered as proper markdown. The user also wants Copy and Download buttons for easy reuse.

## Changes

### 1. Install `react-markdown` — `app/`

```bash
npm install react-markdown
```

No additional plugins needed — `react-markdown` handles headings, lists, code blocks, tables, bold/italic, and links out of the box.

### 2. Create `PlanViewer` component — `app/src/components/ui/plan-viewer.tsx`

A new component that:
- Renders markdown via `react-markdown` with Tailwind prose-like styling scoped to the dark theme
- Has a sticky toolbar at the top with **Copy** and **Download** buttons
- Copy: uses `navigator.clipboard.writeText(plan)`, shows "Copied!" briefly
- Download: creates a Blob and triggers a `.md` file download (filename: `plan.md`)
- Reuses the existing `Button` component (`ghost` variant, `sm` size)

**Markdown styling** — apply Tailwind classes via `react-markdown`'s `components` prop to match the dark theme (e.g., headings get `text-foreground font-semibold`, code blocks get `bg-surface-light rounded border border-border`, links get `text-accent-light`).

### 3. Update checkpoint detail page — `app/src/app/(dashboard)/[owner]/[repo]/[checkpointId]/page.tsx`

Replace the plain-text plan rendering with `<PlanViewer plan={plan} />`.

```diff
+ import { PlanViewer } from "@/components/ui/plan-viewer";
  ...
  {activeTab === "plan" &&
    (planLoading ? (
      <Skeleton className="h-64" />
    ) : plan ? (
-     <div className="rounded-xl border border-border bg-surface p-5">
-       <div className="whitespace-pre-wrap text-sm text-foreground/90 leading-relaxed">
-         {plan}
-       </div>
-     </div>
+     <PlanViewer plan={plan} />
    ) : (
      ...
    ))}
```

## Files

| File | Change |
|------|--------|
| `app/package.json` | Add `react-markdown` dependency |
| `app/src/components/ui/plan-viewer.tsx` | **New** — markdown renderer with copy/download toolbar |
| `app/src/app/(dashboard)/[owner]/[repo]/[checkpointId]/page.tsx` | Import `PlanViewer`, replace plain-text rendering |

## Verification

1. `npm run build` in `app/` — no type errors
2. Visit a checkpoint with a plan — headings, code blocks, lists, tables render as formatted markdown
3. Click **Copy** — plan markdown is copied to clipboard, button briefly shows "Copied!"
4. Click **Download** — browser downloads a `plan.md` file with the raw markdown content
