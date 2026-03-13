import { render, screen } from "@testing-library/svelte";
import { describe, expect, it } from "vitest";
import MetricCard from "$lib/components/MetricCard.svelte";

describe("MetricCard", () => {
  it("shows metric value and change text", () => {
    render(MetricCard, {
      props: {
        metric: {
          label: "Pending Approvals",
          value: "7",
          change: "+1 today",
          direction: "up"
        }
      }
    });

    expect(screen.getByText("Pending Approvals")).toBeTruthy();
    expect(screen.getByText("7")).toBeTruthy();
    expect(screen.getByText("+1 today")).toBeTruthy();
  });
});
