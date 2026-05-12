from typing import Optional
from fastapi import FastAPI
from pydantic import BaseModel

app = FastAPI(title="ml-runner")


class RunRequest(BaseModel):
    action_run_id: int
    config_data: Optional[dict] = None
    event_payload: Optional[dict] = None


@app.get("/health")
def health():
    return {"status": "ok"}


@app.post("/run/{runner_name}")
def run(runner_name: str, body: RunRequest):
    return {
        "status": "success",
        "runner": runner_name,
        "result": {
            "echoed_action_run_id": body.action_run_id,
            "echoed_config_data": body.config_data,
            "echoed_event_payload": body.event_payload,
        },
    }
