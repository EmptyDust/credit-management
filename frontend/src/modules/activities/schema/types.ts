export type ActivitySchemaFieldType =
  | "text"
  | "textarea"
  | "number"
  | "date"
  | "select"
  | "radio"
  | "checkbox"
  | "boolean"

export type ActivitySchemaOperator =
  | "eq"
  | "neq"
  | "in"
  | "not_in"
  | "gt"
  | "gte"
  | "lt"
  | "lte"
  | "contains"
  | "not_empty"

export type ActivitySchemaValue =
  | string
  | number
  | boolean
  | null
  | undefined
  | Array<string | number | boolean>

export interface ActivitySchemaOption {
  value: string
  label: string
  disabled?: boolean
}

export interface ActivitySchemaCondition {
  field: string
  op: ActivitySchemaOperator
  value?: ActivitySchemaValue
}

export interface ActivitySchemaCompareRule {
  type: "compare"
  left: string
  op: "eq" | "neq" | "gt" | "gte" | "lt" | "lte"
  right: string
  message: string
}

export interface ActivitySchemaFieldRules {
  visible_when?: ActivitySchemaCondition[]
  enabled_when?: ActivitySchemaCondition[]
  validate?: ActivitySchemaCompareRule[]
}

export interface ActivitySchemaField {
  key: string
  label: string
  type: ActivitySchemaFieldType
  required?: boolean
  default_value?: ActivitySchemaValue
  placeholder?: string
  options?: ActivitySchemaOption[]
  rules?: ActivitySchemaFieldRules
  display_order?: number
  description?: string
}

export interface ActivitySchema {
  version: number
  fields: ActivitySchemaField[]
}

export interface ActivityCategoryDescriptor {
  key: string
  label: string
}

export interface ActivityCategorySchemaEnvelope {
  category: ActivityCategoryDescriptor
  schema: ActivitySchema
}

export type ActivitySchemaValues = Record<string, ActivitySchemaValue>
export type ActivitySchemaNormalizedValues = Record<string, unknown>

export interface ActivitySchemaFieldValidationResult {
  fieldKey: string
  messages: string[]
}
