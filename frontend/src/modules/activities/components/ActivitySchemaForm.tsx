import { useEffect } from "react"
import { useForm, useWatch } from "react-hook-form"

import { Button } from "@/components/ui/button"
import {
  Form,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form"
import { cn } from "@/lib/utils"
import ActivityDynamicField from "@/modules/activities/components/ActivityDynamicField"
import {
  evaluateActivityFieldEnabled,
  evaluateActivityFieldVisibility,
  getActivitySchemaInitialValues,
  normalizeActivitySchemaValues,
  sortActivitySchemaFields,
  validateActivityField,
} from "@/modules/activities/schema/rule-evaluator"
import type {
  ActivitySchema,
  ActivitySchemaNormalizedValues,
  ActivitySchemaValues,
} from "@/modules/activities/schema/types"

interface ActivitySchemaFormProps {
  schema: ActivitySchema
  initialValues?: ActivitySchemaValues
  submitLabel?: string
  disabled?: boolean
  className?: string
  onSubmit: (values: ActivitySchemaNormalizedValues) => Promise<void> | void
  onValuesChange?: (values: ActivitySchemaValues) => void
}

export default function ActivitySchemaForm({
  schema,
  initialValues,
  submitLabel = "保存",
  disabled = false,
  className,
  onSubmit,
  onValuesChange,
}: ActivitySchemaFormProps) {
  const form = useForm<ActivitySchemaValues>({
    defaultValues: getActivitySchemaInitialValues(schema, initialValues),
    mode: "onSubmit",
    reValidateMode: "onChange",
  })

  const watchedValues = useWatch({
    control: form.control,
  }) as ActivitySchemaValues
  const currentValues = watchedValues || form.getValues()
  const sortedFields = sortActivitySchemaFields(schema.fields)

  useEffect(() => {
    form.reset(getActivitySchemaInitialValues(schema, initialValues))
  }, [form, initialValues, schema])

  useEffect(() => {
    onValuesChange?.(currentValues)

    if (form.formState.submitCount > 0) {
      void form.trigger()
    }
  }, [currentValues, form, onValuesChange])

  const handleSubmit = form.handleSubmit(async (values) => {
    await onSubmit(normalizeActivitySchemaValues(schema, values))
  })

  return (
    <Form {...form}>
      <form className={cn("space-y-6", className)} onSubmit={handleSubmit}>
        {sortedFields.length === 0 ? (
          <div className="rounded-md border border-dashed px-4 py-6 text-sm text-muted-foreground">
            当前分类没有额外字段。
          </div>
        ) : null}

        {sortedFields.map((schemaField) => {
          const visible = evaluateActivityFieldVisibility(schemaField, currentValues)

          if (!visible) {
            return null
          }

          const enabled =
            !disabled && evaluateActivityFieldEnabled(schemaField, currentValues)

          return (
            <FormField
              control={form.control}
              key={schemaField.key}
              name={schemaField.key}
              rules={{
                validate: () => {
                  const errors = validateActivityField(schemaField, form.getValues())
                  return errors[0] ?? true
                },
              }}
              render={({ field }) => (
                <FormItem>
                  <FormLabel>
                    {schemaField.label}
                    {schemaField.required ? (
                      <span className="ml-1 text-destructive">*</span>
                    ) : null}
                  </FormLabel>
                  <ActivityDynamicField
                    disabled={!enabled}
                    name={schemaField.key}
                    onBlur={field.onBlur}
                    onChange={field.onChange}
                    schemaField={schemaField}
                    value={field.value}
                  />
                  {schemaField.description ? (
                    <FormDescription>{schemaField.description}</FormDescription>
                  ) : null}
                  <FormMessage />
                </FormItem>
              )}
            />
          )
        })}

        <div className="flex justify-end">
          <Button disabled={disabled || form.formState.isSubmitting} type="submit">
            {submitLabel}
          </Button>
        </div>
      </form>
    </Form>
  )
}
