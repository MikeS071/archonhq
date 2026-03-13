# reduction

Reduction service for advanced workload merge decisions.

Implemented in M7:
- supported merge strategy registry (`append_only_v1`, `key_upsert_v1`, `section_patch_v1`, `ast_patch_v1`, `topk_rank_v1`, `quorum_fact_v1`, `reduce_tree_v1`)
- deterministic candidate ranking
- merge decision output with status, winner, ranked candidates, merged patch ops, and output state reference
