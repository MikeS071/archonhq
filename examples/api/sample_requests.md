### Create task
POST /v1/tasks
Content-Type: application/json

{
  "workspace_id": "ws_01",
  "task_family": "doc.section.write",
  "title": "Write threat model section",
  "description": "Draft section 3 using evidence pack",
  "input_refs": ["art_outline_01", "art_sources_01"]
}

### Submit result
POST /v1/results
Content-Type: application/json

{
  "result_id": "res_01",
  "task_id": "task_01",
  "lease_id": "lease_01",
  "output_refs": ["art_01"]
}
