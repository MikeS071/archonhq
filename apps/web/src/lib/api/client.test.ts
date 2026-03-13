import { afterEach, describe, expect, it, vi } from "vitest";
import { ApiError, apiGet } from "$lib/api/client";

function jsonResponse(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "content-type": "application/json" }
  });
}

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe("apiGet", () => {
  it("returns parsed JSON for successful requests", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => jsonResponse({ ok: true })));

    const result = await apiGet<{ ok: boolean }>("/v1/tasks/feed");
    expect(result.ok).toBe(true);
  });

  it("throws ApiError with envelope fields", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        jsonResponse(
          {
            error: {
              code: "forbidden",
              message: "No access",
              correlation_id: "corr_123"
            }
          },
          403
        )
      )
    );

    await expect(apiGet("/v1/tasks/feed")).rejects.toBeInstanceOf(ApiError);
  });
});
