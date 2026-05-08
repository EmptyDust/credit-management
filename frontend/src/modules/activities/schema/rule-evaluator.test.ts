import { describe, expect, it } from "vitest";

import {
  evaluateActivityFieldEnabled,
  evaluateActivityFieldVisibility,
  validateActivityFieldRules,
} from "./rule-evaluator";

describe("activity schema rule evaluator", () => {
  it("evaluates visible_when conditions", () => {
    const field = {
      key: "location",
      rules: {
        visible_when: [
          {
            field: "activity_mode",
            op: "eq",
            value: "offline",
          },
        ],
      },
    };

    expect(
      evaluateActivityFieldVisibility(field, {
        activity_mode: "offline",
      })
    ).toBe(true);

    expect(
      evaluateActivityFieldVisibility(field, {
        activity_mode: "online",
      })
    ).toBe(false);
  });

  it("evaluates enabled_when conditions", () => {
    const field = {
      key: "approval_reason",
      rules: {
        enabled_when: [
          {
            field: "approval_needed",
            op: "eq",
            value: true,
          },
        ],
      },
    };

    expect(
      evaluateActivityFieldEnabled(field, {
        approval_needed: true,
      })
    ).toBe(true);

    expect(
      evaluateActivityFieldEnabled(field, {
        approval_needed: false,
      })
    ).toBe(false);
  });

  it("validates compare rules", () => {
    const field = {
      key: "end_date",
      label: "End Date",
      rules: {
        validate: [
          {
            type: "compare",
            left: "end_date",
            op: "gte",
            right: "start_date",
            message: "End date cannot be earlier than start date",
          },
        ],
      },
    };

    expect(
      validateActivityFieldRules(field, {
        start_date: "2026-05-10",
        end_date: "2026-05-12",
      })
    ).toEqual([]);

    expect(
      validateActivityFieldRules(field, {
        start_date: "2026-05-10",
        end_date: "2026-05-01",
      })
    ).toEqual(["End date cannot be earlier than start date"]);
  });

  it("supports in conditions", () => {
    const field = {
      key: "audience",
      rules: {
        visible_when: [
          {
            field: "category",
            op: "in",
            value: ["training", "exchange"],
          },
        ],
      },
    };

    expect(
      evaluateActivityFieldVisibility(field, {
        category: "training",
      })
    ).toBe(true);

    expect(
      evaluateActivityFieldVisibility(field, {
        category: "competition",
      })
    ).toBe(false);
  });

  it("supports not_empty conditions", () => {
    const field = {
      key: "approval_note",
      rules: {
        enabled_when: [
          {
            field: "approval_comment",
            op: "not_empty",
          },
        ],
      },
    };

    expect(
      evaluateActivityFieldEnabled(field, {
        approval_comment: "needs review",
      })
    ).toBe(true);

    expect(
      evaluateActivityFieldEnabled(field, {
        approval_comment: "",
      })
    ).toBe(false);
  });
});
