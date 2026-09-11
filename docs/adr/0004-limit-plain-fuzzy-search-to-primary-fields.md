---
type: ADR
id: "0004"
title: "Limit plain fuzzy search to primary fields"
status: active
date: 2026-09-11
supersedes: "0002"
---

## Context

Plain fuzzy search examined every issue text field. Short queries therefore matched letters in descriptions, notes, comments, and references to other issue IDs. The resulting rows often had no visible title, ID, or label that explained the match. Searching for a short issue path such as `ljw` could show many issues whose notes merely mentioned the target as a blocker or related task.

## Decision

**Plain fuzzy search and its Tab completions use only issue IDs, titles, and labels. Other metadata remains available through structured predicates.**

Descriptions, design, acceptance criteria, notes, comments, dependency prose, status, priority, type, assignee, project, and external references do not participate in a plain query. Structured predicates such as `status:open`, `assignee:andre`, and `project:b9s` continue to target their fields explicitly.

## Options considered

- **ID, title, and labels** (chosen): keeps ordinary search explainable from visible identity data while retaining fuzzy abbreviations and label discovery.
- **All issue text**: maximizes recall, but short queries produce many invisible and accidental matches.
- **ID and title only**: gives the highest precision, but breaks the established workflow of finding issues through labels such as `lane-attempt=1`.
- **Rank all issue text by relevance**: could expose deep content without flooding the result set, but requires ranking and a different result contract.

## Consequences

Short issue paths and title fragments return a small, understandable result set. Mentions in blocker notes, descriptions, or comments no longer make an issue appear. Label abbreviations remain supported, including `lna1` for `lane-attempt=1`. Tab completion cannot suggest a hidden term that plain search would reject.

Users must use a structured predicate for supported secondary metadata. Deep content search may be reconsidered as a separate ranked mode if users need it without sacrificing the precision of ordinary `/` search.
