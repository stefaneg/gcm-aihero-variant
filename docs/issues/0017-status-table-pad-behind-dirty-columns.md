# 0017 — Status table: pad `behind=` and `dirty=` to max width across rows

Status: ready-for-agent

## What to build

The **Status Table** loses column alignment when any row has a multi-digit `behind` or `dirty` count. The first two columns (derived path, branch) are width-padded across rows, but the `behind=%d` and `dirty=%d` fields are printed raw, so a row with `behind=408` shifts every column to its right (including the badge cluster) compared to a row with `behind=0`. The badges then no longer form a clean vertical band, which is the whole point of having them as a column.

Fix: compute the maximum digit width of `behind` and the maximum digit width of `dirty` across the rendered (filter-applied) row set, then format each numeric value **right-aligned** to that width. The `=` stays glued to its label; only the number is padded on the left. Example with a 408 in the set:

```
behind=  0  dirty=0
behind=408  dirty=0
behind=  9  dirty=1
```

Right-alignment is chosen so that magnitudes are visually comparable down the column — numbers conventionally right-align, even though the textual columns to the left use left-alignment.

The fix applies regardless of which filter is active (`gcm status`, `gcm status --non-default`); the widths are computed from the rows actually being rendered, not from the unfiltered set. Empty-table and filter-empty-table cases (covered by 0014) are unaffected because no rows are rendered.

Badge columns naturally re-align as soon as the numeric columns have stable widths — no separate badge-padding work is needed.

## Acceptance criteria

- [ ] When all rows have single-digit `behind` and `dirty` counts, table output is byte-for-byte unchanged from the current behaviour (modulo any leading space introduced by the new width — see next criterion)
- [ ] When at least one row has a multi-digit `behind` or `dirty`, all rows in that column are right-padded to the same width, and the columns following them (including any badges) align vertically
- [ ] Numeric values are right-aligned within their padded width (`behind=  9`, not `behind=9  `)
- [ ] Width is computed from the rendered row set, so `--non-default` and other filters produce a table sized to what they render, not to the unfiltered set
- [ ] Unit tests cover: mixed single- and multi-digit `behind` values in the same table, mixed `dirty` values, both columns multi-digit simultaneously
- [ ] Existing alignment-related tests in `formatter_test.go` are updated where the new padding changes their expected output, with the multi-digit cases added rather than replacing the existing single-digit assertions

## Blocked by

- None - can start immediately
