import type {
  ActivitySchema,
  ActivitySchemaCondition,
  ActivitySchemaField,
  ActivitySchemaFieldValidationResult,
  ActivitySchemaNormalizedValues,
  ActivitySchemaValue,
  ActivitySchemaValues,
} from "./types"

const ISO_DATE_PATTERN = /^\d{4}-\d{2}-\d{2}$/
const NUMBER_PATTERN = /^-?\d+(\.\d+)?$/

function isArrayValue(value: ActivitySchemaValue): value is Array<string | number | boolean> {
  return Array.isArray(value)
}

function isEmptyValue(value: ActivitySchemaValue) {
  if (value === null || value === undefined) {
    return true
  }

  if (typeof value === "string") {
    return value.trim() === ""
  }

  if (Array.isArray(value)) {
    return value.length === 0
  }

  return false
}

function normalizeComparableValue(value: ActivitySchemaValue) {
  if (value === null || value === undefined || Array.isArray(value)) {
    return null
  }

  if (typeof value === "number") {
    return { kind: "number" as const, value }
  }

  if (typeof value === "boolean") {
    return { kind: "boolean" as const, value }
  }

  if (ISO_DATE_PATTERN.test(value)) {
    return { kind: "date" as const, value }
  }

  if (NUMBER_PATTERN.test(value)) {
    return { kind: "number" as const, value: Number(value) }
  }

  return { kind: "string" as const, value }
}

function valuesEqual(left: ActivitySchemaValue, right: ActivitySchemaValue): boolean {
  if (Array.isArray(left) || Array.isArray(right)) {
    if (!Array.isArray(left) || !Array.isArray(right) || left.length !== right.length) {
      return false
    }

    return left.every((item, index) => valuesEqual(item, right[index]))
  }

  const leftComparable = normalizeComparableValue(left)
  const rightComparable = normalizeComparableValue(right)

  if (leftComparable && rightComparable && leftComparable.kind === rightComparable.kind) {
    return leftComparable.value === rightComparable.value
  }

  return left === right
}

function compareOrderedValues(
  left: ActivitySchemaValue,
  right: ActivitySchemaValue,
  op: "gt" | "gte" | "lt" | "lte"
) {
  const leftComparable = normalizeComparableValue(left)
  const rightComparable = normalizeComparableValue(right)

  if (!leftComparable || !rightComparable) {
    return false
  }

  if (leftComparable.kind === "boolean" || rightComparable.kind === "boolean") {
    return false
  }

  if (leftComparable.kind !== rightComparable.kind) {
    return false
  }

  switch (op) {
    case "gt":
      return leftComparable.value > rightComparable.value
    case "gte":
      return leftComparable.value >= rightComparable.value
    case "lt":
      return leftComparable.value < rightComparable.value
    case "lte":
      return leftComparable.value <= rightComparable.value
  }
}

function asScalarValues(value: ActivitySchemaValue): Array<string | number | boolean> {
  if (Array.isArray(value)) {
    return value
  }

  if (value === null || value === undefined) {
    return []
  }

  return [value]
}

function evaluateActivityCondition(
  condition: ActivitySchemaCondition,
  values: ActivitySchemaValues
) {
  const actual = values[condition.field]
  const expected = condition.value

  switch (condition.op) {
    case "eq":
      return valuesEqual(actual, expected)
    case "neq":
      return !valuesEqual(actual, expected)
    case "in": {
      const actualValues = asScalarValues(actual)
      const expectedValues = asScalarValues(expected)
      return actualValues.some((item) =>
        expectedValues.some((candidate) => valuesEqual(item, candidate))
      )
    }
    case "not_in": {
      const actualValues = asScalarValues(actual)
      const expectedValues = asScalarValues(expected)
      return actualValues.every(
        (item) => !expectedValues.some((candidate) => valuesEqual(item, candidate))
      )
    }
    case "contains":
      if (Array.isArray(actual)) {
        return asScalarValues(expected).some((candidate) =>
          actual.some((item) => valuesEqual(item, candidate))
        )
      }

      if (typeof actual === "string") {
        return asScalarValues(expected).some((candidate) =>
          actual.includes(String(candidate))
        )
      }

      return false
    case "not_empty":
      return !isEmptyValue(actual)
    case "gt":
    case "gte":
    case "lt":
    case "lte":
      return compareOrderedValues(actual, expected, condition.op)
    default:
      return false
  }
}

function evaluateActivityConditions(
  conditions: ActivitySchemaCondition[] | undefined,
  values: ActivitySchemaValues
) {
  if (!conditions || conditions.length === 0) {
    return true
  }

  return conditions.every((condition) => evaluateActivityCondition(condition, values))
}

function normalizeActivityFieldValue(
  field: ActivitySchemaField,
  rawValue: ActivitySchemaValue
): unknown {
  if (field.type === "checkbox") {
    if (!Array.isArray(rawValue)) {
      return undefined
    }

    const selectedValues = rawValue.filter(
      (value): value is string | number | boolean =>
        value !== null && value !== undefined && `${value}`.trim() !== ""
    )

    return selectedValues.length > 0 ? selectedValues : undefined
  }

  if (field.type === "boolean") {
    return rawValue === true
  }

  if (field.type === "number") {
    if (rawValue === "" || rawValue === null || rawValue === undefined) {
      return undefined
    }

    if (typeof rawValue === "number") {
      return Number.isFinite(rawValue) ? rawValue : undefined
    }

    if (typeof rawValue === "string") {
      const numericValue = Number(rawValue)
      return Number.isFinite(numericValue) ? numericValue : undefined
    }

    return undefined
  }

  if (typeof rawValue === "string") {
    return rawValue.trim() === "" ? undefined : rawValue
  }

  return rawValue
}

function validateAllowedOptions(field: ActivitySchemaField, value: ActivitySchemaValue) {
  const options = field.options ?? []

  if (options.length === 0 || value === null || value === undefined) {
    return []
  }

  const allowedValues = new Set(options.map((option) => option.value))

  if (Array.isArray(value)) {
    return value.every((item) => allowedValues.has(String(item)))
      ? []
      : [`${field.label}包含无效选项`]
  }

  if (typeof value === "string" && value.trim() === "") {
    return []
  }

  return allowedValues.has(String(value)) ? [] : [`${field.label}包含无效选项`]
}

export function sortActivitySchemaFields(fields: ActivitySchemaField[]) {
  return [...fields].sort((left, right) => {
    const orderDiff = (left.display_order ?? 0) - (right.display_order ?? 0)

    if (orderDiff !== 0) {
      return orderDiff
    }

    return left.key.localeCompare(right.key)
  })
}

export function getActivitySchemaInitialValues(
  schema: ActivitySchema,
  initialValues: ActivitySchemaValues = {}
) {
  return sortActivitySchemaFields(schema.fields).reduce<ActivitySchemaValues>((accumulator, field) => {
    if (Object.prototype.hasOwnProperty.call(initialValues, field.key)) {
      accumulator[field.key] = initialValues[field.key]
      return accumulator
    }

    if (field.default_value !== undefined) {
      accumulator[field.key] = field.default_value
      return accumulator
    }

    if (field.type === "checkbox") {
      accumulator[field.key] = []
      return accumulator
    }

    if (field.type === "boolean") {
      accumulator[field.key] = false
      return accumulator
    }

    accumulator[field.key] = ""
    return accumulator
  }, {})
}

export function evaluateActivityFieldVisibility(
  field: Pick<ActivitySchemaField, "rules">,
  values: ActivitySchemaValues
) {
  return evaluateActivityConditions(field.rules?.visible_when, values)
}

export function evaluateActivityFieldEnabled(
  field: Pick<ActivitySchemaField, "rules">,
  values: ActivitySchemaValues
) {
  return evaluateActivityConditions(field.rules?.enabled_when, values)
}

export function validateActivityFieldRules(
  field: Pick<ActivitySchemaField, "rules">,
  values: ActivitySchemaValues
) {
  if (!evaluateActivityFieldVisibility(field, values)) {
    return []
  }

  const messages: string[] = []

  for (const rule of field.rules?.validate ?? []) {
    const left = values[rule.left]
    const right = values[rule.right]

    if (isEmptyValue(left) || isEmptyValue(right)) {
      continue
    }

    let valid = false
    switch (rule.op) {
      case "eq":
        valid = valuesEqual(left, right)
        break
      case "neq":
        valid = !valuesEqual(left, right)
        break
      case "gt":
      case "gte":
      case "lt":
      case "lte":
        valid = compareOrderedValues(left, right, rule.op)
        break
    }

    if (!valid) {
      messages.push(rule.message)
    }
  }

  return messages
}

export function validateActivityField(
  field: ActivitySchemaField,
  values: ActivitySchemaValues
) {
  if (!evaluateActivityFieldVisibility(field, values)) {
    return []
  }

  const messages: string[] = []
  const currentValue = values[field.key]

  if (field.required && isEmptyValue(currentValue)) {
    messages.push(`${field.label}不能为空`)
  }

  messages.push(...validateAllowedOptions(field, currentValue))
  messages.push(...validateActivityFieldRules(field, values))

  return messages
}

export function validateActivitySchemaValues(
  schema: ActivitySchema,
  values: ActivitySchemaValues
) {
  const results: ActivitySchemaFieldValidationResult[] = []

  for (const field of sortActivitySchemaFields(schema.fields)) {
    const messages = validateActivityField(field, values)

    if (messages.length > 0) {
      results.push({
        fieldKey: field.key,
        messages,
      })
    }
  }

  return results
}

export function normalizeActivitySchemaValues(
  schema: ActivitySchema,
  values: ActivitySchemaValues
) {
  return sortActivitySchemaFields(schema.fields).reduce<ActivitySchemaNormalizedValues>(
    (accumulator, field) => {
      if (!evaluateActivityFieldVisibility(field, values)) {
        return accumulator
      }

      const normalizedValue = normalizeActivityFieldValue(field, values[field.key])

      if (normalizedValue !== undefined) {
        accumulator[field.key] = normalizedValue
      }

      return accumulator
    },
    {}
  )
}
