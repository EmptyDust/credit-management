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

// 简单内存缓存，避免多个页面/组件重复请求相同的 options
let cachedOptions: OptionsResponse | null = null;
let optionsPromise: Promise<OptionsResponse> | null = null;

export async function getOptions(): Promise<OptionsResponse> {
  if (cachedOptions) return cachedOptions;
  if (optionsPromise) return optionsPromise;

  optionsPromise = (async () => {
    const res = await apiClient.get("/config/options");
    const data = res.data?.data || {};
    const normalized: OptionsResponse = {
      colleges: Array.isArray(data.colleges) ? data.colleges : [],
      majors:
        typeof data.majors === "object" && data.majors !== null
          ? data.majors
          : {},
      classes:
        typeof data.classes === "object" && data.classes !== null
          ? data.classes
          : {},
      grades: Array.isArray(data.grades) ? data.grades : [],
      user_statuses: Array.isArray(data.user_statuses)
        ? data.user_statuses
        : [],
      teacher_titles: Array.isArray(data.teacher_titles)
        ? data.teacher_titles
        : [],
    };
    cachedOptions = normalized;
    optionsPromise = null;
    return normalized;
  })();

  return optionsPromise;
}

export type ActivityOptions = {
  categories: SelectOption[];
  statuses: SelectOption[];
  review_actions: SelectOption[];
  category_fields?: Record<
    string,
    Array<{
      name: string;
      label: string;
      type: string;
      required?: boolean;
      options?: SelectOption[];
      min?: number;
      max?: number;
      maxLength?: number;
    }>
  >;
};

// 活动相关 options 的缓存，避免在列表页、详情页、弹窗等多处重复请求
let cachedActivityOptions: ActivityOptions | null = null;
let activityOptionsPromise: Promise<ActivityOptions> | null = null;

export async function getActivityOptions(): Promise<ActivityOptions> {
  if (cachedActivityOptions) return cachedActivityOptions;
  if (activityOptionsPromise) return activityOptionsPromise;

  activityOptionsPromise = (async () => {
    const res = await apiClient.get("/activities/config/options");
    const data = res.data?.data || {};
    const normalized: ActivityOptions = {
      categories: Array.isArray(data.categories) ? data.categories : [],
      statuses: Array.isArray(data.statuses) ? data.statuses : [],
      review_actions: Array.isArray(data.review_actions)
        ? data.review_actions
        : [],
      category_fields: data.category_fields || {},
    };
    cachedActivityOptions = normalized;
    activityOptionsPromise = null;
    return normalized;
  })();

  return activityOptionsPromise;
}

