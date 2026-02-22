# Show GitHub Username & Logo for Human Messages in Transcript

## Context

In the transcript viewer on the checkpoint page, human messages currently display a generic person icon avatar and the label "You". The user wants to replace these with the actual GitHub username and a GitHub logo, to better identify who wrote the messages.

The GitHub username (`session.user.login`) and avatar URL (`session.user.image`) are already available via NextAuth's session. The transcript viewer just needs to receive and use them.

## Changes

### 1. Thread `userName` through `TranscriptViewer` — `app/src/app/(dashboard)/[owner]/[repo]/[checkpointId]/page.tsx`

Use NextAuth's `useSession()` hook to get the logged-in user's GitHub login, and pass it as a new `userName` prop to `<TranscriptViewer>`.

```diff
+ import { useSession } from "next-auth/react";
  ...
  export default function CheckpointDetailPage(...) {
+   const { data: session } = useSession();
    ...
    <TranscriptViewer
      messages={messages || []}
      agentName={checkpoint.agent || undefined}
      agentPercent={checkpoint.agent_percent}
+     userName={session?.user?.login || undefined}
    />
```

### 2. Accept & pass `userName` in transcript components — `app/src/components/ui/transcript-viewer.tsx`

- Add `userName?: string` to `TranscriptViewerProps`, `SessionAccordion` props, and `MessageRow` props
- Thread it from `TranscriptViewer` → `SessionAccordion` → `MessageRow`
- In `MessageRow`, replace the "You" label with `userName ?? "You"`
- Replace `HumanAvatar` (generic person icon) with a GitHub logo SVG

**`HumanAvatar` replacement** — swap the person icon for the GitHub mark:

```tsx
function GitHubAvatar() {
  return (
    <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-surface-light border border-border">
      <svg
        width="16"
        height="16"
        viewBox="0 0 16 16"
        fill="currentColor"
        className="text-muted"
      >
        <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z" />
      </svg>
    </div>
  );
}
```

**`MessageRow` label change:**

```diff
- {isHuman ? <HumanAvatar /> : <AssistantAvatar />}
+ {isHuman ? <GitHubAvatar /> : <AssistantAvatar />}
  ...
  <span className="text-xs font-medium text-foreground">
-   {isHuman ? "You" : "Assistant"}
+   {isHuman ? (userName ?? "You") : "Assistant"}
  </span>
```

## Files Modified

| File | Change |
|------|--------|
| `app/src/app/(dashboard)/[owner]/[repo]/[checkpointId]/page.tsx` | Import `useSession`, pass `userName` to `TranscriptViewer` |
| `app/src/components/ui/transcript-viewer.tsx` | Add `userName` prop, rename `HumanAvatar` → `GitHubAvatar` with GitHub SVG, display `userName` instead of "You" |

## Verification

1. `npm run build` in `app/` — no type errors
2. Visit a checkpoint page while logged in — human messages show your GitHub username and the GitHub logo instead of "You" with a person icon
3. If session hasn't loaded yet, falls back to "You"
