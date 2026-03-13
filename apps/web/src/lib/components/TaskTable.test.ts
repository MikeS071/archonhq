import { render, screen } from "@testing-library/svelte";
import { describe, expect, it } from "vitest";
import TaskTable from "$lib/components/TaskTable.svelte";
import { tasks } from "$lib/mock-data";

describe("TaskTable", () => {
  it("renders task rows and detail links", () => {
    const subset = tasks.slice(0, 2);
    render(TaskTable, { props: { tasks: subset } });

    expect(screen.getByText(subset[0].title)).toBeTruthy();
    expect(screen.getByText(subset[1].family)).toBeTruthy();

    const openLinks = screen.getAllByRole("link", { name: "Open" });
    expect(openLinks).toHaveLength(2);
    expect(openLinks[0].getAttribute("href")).toBe(`/tasks/${subset[0].id}`);
  });
});
