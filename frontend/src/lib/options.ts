import apiClient from "./api";

export type SelectOption = { value: string; label: string };

export type OptionsResponse = {
  colleges: SelectOption[];
  majors: Record<string, SelectOption[]>;
  classes: Record<string, SelectOption[]>;
  grades: SelectOption[];
  user_statuses: SelectOption[];
  teacher_titles: SelectOption[];
};

export async function getOptions(): Promise<OptionsResponse> {
  const res = await apiClient.get("/config/options");
  const data = res.data?.data || {};
  return {
    colleges: Array.isArray(data.colleges) ? data.colleges : [],
    majors: typeof data.majors === 'object' && data.majors !== null ? data.majors : {},
    classes: typeof data.classes === 'object' && data.classes !== null ? data.classes : {},
    grades: Array.isArray(data.grades) ? data.grades : [],
    user_statuses: Array.isArray(data.user_statuses) ? data.user_statuses : [],
    teacher_titles: Array.isArray(data.teacher_titles) ? data.teacher_titles : [],
  };
}

export type ActivityOptions = {
  categories: SelectOption[];
  statuses: SelectOption[];
  review_actions: SelectOption[];
  category_fields?: Record<string, Array<{ name: string; label: string; type: string; required?: boolean; options?: SelectOption[]; min?: number; max?: number; maxLength?: number }>>;
};

export async function getActivityOptions(): Promise<ActivityOptions> {
  const res = await apiClient.get("/activities/config/options");
  const data = res.data?.data || {};
  return {
    categories: Array.isArray(data.categories) ? data.categories : [],
    statuses: Array.isArray(data.statuses) ? data.statuses : [],
    review_actions: Array.isArray(data.review_actions) ? data.review_actions : [],
    category_fields: data.category_fields || {},
  };
}


