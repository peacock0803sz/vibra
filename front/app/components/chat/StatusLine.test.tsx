import { render, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import type { StatusLineData } from "~/lib/hooks/useStatusLine";

import { StatusLine } from "./StatusLine";

const initialData: StatusLineData = {
  hostname: "-",
  repository: "-",
  branch: "-",
  agentType: "-",
  modelName: "-",
};

const fullData: StatusLineData = {
  hostname: "myhost",
  repository: "owner/repo",
  branch: "main",
  agentType: "Claude",
  modelName: "claude-code",
};

describe("StatusLine", () => {
  it("displays all labels in initial state", () => {
    const { container } = render(<StatusLine data={initialData} />);
    const view = within(container);
    expect(view.getByText("host")).toBeInTheDocument();
    expect(view.getByText("repo")).toBeInTheDocument();
    expect(view.getByText("branch")).toBeInTheDocument();
    expect(view.getByText("agent")).toBeInTheDocument();
    expect(view.getByText("model")).toBeInTheDocument();
  });

  it("displays '-' for each label value in initial state", () => {
    const { container } = render(<StatusLine data={initialData} />);
    const view = within(container);
    // Verify that the value span next to each label element is '-'
    for (const label of ["host", "repo", "branch", "agent", "model"]) {
      const labelEl = view.getByText(label);
      expect(labelEl.nextElementSibling?.textContent).toBe("-");
    }
  });

  it("displays actual values correctly", () => {
    const { container } = render(<StatusLine data={fullData} />);
    const view = within(container);
    expect(view.getByText("myhost")).toBeInTheDocument();
    expect(view.getByText("owner/repo")).toBeInTheDocument();
    expect(view.getByText("main")).toBeInTheDocument();
    expect(view.getByText("Claude")).toBeInTheDocument();
    expect(view.getByText("claude-code")).toBeInTheDocument();
  });

  it("adds title attribute to long values (repo and branch are truncated)", () => {
    const longData: StatusLineData = {
      ...fullData,
      repository: "very-long-organization/very-long-repository-name",
      branch: "feature/very-long-branch-name-that-exceeds-display-width",
    };
    const { container } = render(<StatusLine data={longData} />);
    const view = within(container);
    expect(view.getByTitle(longData.repository)).toBeInTheDocument();
    expect(view.getByTitle(longData.branch)).toBeInTheDocument();
  });

  it("does not add title attribute to hostname and model (not truncated)", () => {
    const { container } = render(<StatusLine data={fullData} />);
    const view = within(container);
    expect(view.queryByTitle(fullData.hostname)).not.toBeInTheDocument();
    expect(view.queryByTitle(fullData.modelName)).not.toBeInTheDocument();
  });
});
