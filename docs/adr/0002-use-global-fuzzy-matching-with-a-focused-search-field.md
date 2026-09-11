---
type: ADR
id: "0002"
title: "Use global fuzzy matching with a focused search field"
status: superseded
date: 2026-09-11
supersedes: "0001"
superseded_by: "0004"
---

## Context

ADR 0001 unified query ownership across tree, list, and board, but plain text only matched ID and title substrings. String predicates also required complete values. A partial label such as `label:lane-a` therefore returned no issues even when labels such as `lane-attempt=1` were loaded. The persistent query bar mixed search input, filter chips, result counts, completion hints, project scope, and sort state on one line, which made the input difficult to scan.

## Decision

**Plain text fuzzily searches every meaningful text field on an issue, and string predicates use the same fuzzy matcher within their selected field. The shared query state remains root-owned, while its visual surface is a dedicated bordered search field.**

Searchable content includes ID, title, description, design, acceptance criteria, notes, status, priority, type, assignee, labels, project, external reference, and comments. Priority predicates remain exact because fuzzy numeric matching would make `p1` match `p10`.

## Options considered

- **Global fuzzy matching with a focused input** (chosen): makes partial labels and abbreviated terms useful everywhere, with a compact visual surface.
- **Fuzzy plain text with exact predicates**: keeps strict facet semantics, but preserves the partial-label failure when a user selects `label:` completion.
- **Backend full-text search**: can scale beyond in-memory data, but adds backend-dependent behavior and latency to an interactive filter.

## Consequences

The same abbreviated query returns the same result set in tree, list, and board. Queries may match content that is not visible in list columns, so users can find an issue from its description, notes, or comments. Structured predicates remain available when a user wants to constrain the matching field. Search results retain the active view's sort order because matching returns a boolean rather than a relevance rank.

Re-evaluate the in-memory matcher if loaded workspaces grow enough to make per-keystroke scans visible, or if users need relevance-ranked results rather than filtering.
