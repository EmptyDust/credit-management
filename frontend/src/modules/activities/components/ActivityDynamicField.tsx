import { Checkbox } from "@/components/ui/checkbox"
import { FormControl } from "@/components/ui/form"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Textarea } from "@/components/ui/textarea"
import { cn } from "@/lib/utils"
import type {
  ActivitySchemaField,
  ActivitySchemaValue,
} from "@/modules/activities/schema/types"

interface ActivityDynamicFieldProps {
  name: string
  schemaField: ActivitySchemaField
  value: ActivitySchemaValue
  disabled?: boolean
  onBlur?: () => void
  onChange: (value: ActivitySchemaValue) => void
}

function asString(value: ActivitySchemaValue) {
  return typeof value === "string" ? value : ""
}

function asCheckboxValues(value: ActivitySchemaValue) {
  return Array.isArray(value)
    ? value.filter((item): item is string => typeof item === "string")
    : []
}

function toggleCheckboxValue(
  selectedValues: string[],
  nextValue: string,
  checked: boolean
) {
  if (checked) {
    return [...selectedValues, nextValue]
  }

  return selectedValues.filter((value) => value !== nextValue)
}

export default function ActivityDynamicField({
  name,
  schemaField,
  value,
  disabled = false,
  onBlur,
  onChange,
}: ActivityDynamicFieldProps) {
  switch (schemaField.type) {
    case "textarea":
      return (
        <FormControl>
          <Textarea
            disabled={disabled}
            onBlur={onBlur}
            onChange={(event) => onChange(event.target.value)}
            placeholder={schemaField.placeholder}
            value={asString(value)}
          />
        </FormControl>
      )
    case "number":
      return (
        <FormControl>
          <Input
            disabled={disabled}
            inputMode="decimal"
            onBlur={onBlur}
            onChange={(event) => onChange(event.target.value)}
            placeholder={schemaField.placeholder}
            type="number"
            value={typeof value === "number" ? value : asString(value)}
          />
        </FormControl>
      )
    case "date":
      return (
        <FormControl>
          <Input
            disabled={disabled}
            onBlur={onBlur}
            onChange={(event) => onChange(event.target.value)}
            type="date"
            value={asString(value)}
          />
        </FormControl>
      )
    case "select":
      return (
        <Select
          disabled={disabled}
          onValueChange={(nextValue) => onChange(nextValue)}
          value={asString(value) || undefined}
        >
          <FormControl>
            <SelectTrigger>
              <SelectValue
                placeholder={schemaField.placeholder || `请选择${schemaField.label}`}
              />
            </SelectTrigger>
          </FormControl>
          <SelectContent>
            {(schemaField.options ?? []).map((option) => (
              <SelectItem disabled={option.disabled} key={option.value} value={option.value}>
                {option.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      )
    case "radio":
      return (
        <FormControl>
          <div className="grid gap-2" role="radiogroup">
            {(schemaField.options ?? []).map((option) => {
              const inputId = `${name}-${option.value}`
              const checked = asString(value) === option.value

              return (
                <Label
                  className={cn(
                    "flex items-center gap-3 rounded-md border px-3 py-2 transition-colors",
                    checked && "border-primary bg-primary/5",
                    (disabled || option.disabled) && "cursor-not-allowed opacity-60"
                  )}
                  htmlFor={inputId}
                  key={option.value}
                >
                  <input
                    checked={checked}
                    className="h-4 w-4"
                    disabled={disabled || option.disabled}
                    id={inputId}
                    name={name}
                    onBlur={onBlur}
                    onChange={() => onChange(option.value)}
                    type="radio"
                    value={option.value}
                  />
                  <span>{option.label}</span>
                </Label>
              )
            })}
          </div>
        </FormControl>
      )
    case "checkbox": {
      const selectedValues = asCheckboxValues(value)

      return (
        <FormControl>
          <div className="grid gap-2">
            {(schemaField.options ?? []).map((option) => {
              const checked = selectedValues.includes(option.value)
              const inputId = `${name}-${option.value}`

              return (
                <Label
                  className={cn(
                    "flex items-center gap-3 rounded-md border px-3 py-2 transition-colors",
                    checked && "border-primary bg-primary/5",
                    (disabled || option.disabled) && "cursor-not-allowed opacity-60"
                  )}
                  htmlFor={inputId}
                  key={option.value}
                >
                  <Checkbox
                    checked={checked}
                    disabled={disabled || option.disabled}
                    onCheckedChange={(nextChecked) =>
                      onChange(
                        toggleCheckboxValue(
                          selectedValues,
                          option.value,
                          nextChecked === true
                        )
                      )
                    }
                  />
                  <span>{option.label}</span>
                </Label>
              )
            })}
          </div>
        </FormControl>
      )
    }
    case "boolean":
      return (
        <FormControl>
          <div className="flex items-center gap-3 rounded-md border px-3 py-3">
            <Checkbox
              checked={value === true}
              disabled={disabled}
              onCheckedChange={(nextChecked) => onChange(nextChecked === true)}
            />
            <span className="text-sm text-muted-foreground">
              {schemaField.placeholder || schemaField.options?.[0]?.label || "启用此项"}
            </span>
          </div>
        </FormControl>
      )
    case "text":
    default:
      return (
        <FormControl>
          <Input
            disabled={disabled}
            onBlur={onBlur}
            onChange={(event) => onChange(event.target.value)}
            placeholder={schemaField.placeholder}
            type="text"
            value={asString(value)}
          />
        </FormControl>
      )
  }
}
