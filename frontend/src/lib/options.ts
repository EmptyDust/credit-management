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
    const [organizationRes, gradesRes, titlesRes] = await Promise.all([
      apiClient.get("/options/organization/tree"),
      apiClient.get("/options/grades"),
      apiClient.get("/options/teacher-titles").catch(() => ({ data: { data: { items: [] } } })),
    ]);

    const treeItems = organizationRes.data?.data?.items || [];
    const colleges: SelectOption[] = [];
    const majors: Record<string, SelectOption[]> = {};
    const classes: Record<string, SelectOption[]> = {};

    for (const school of treeItems) {
      const collegeNodes = Array.isArray(school.children) ? school.children : [];
      for (const college of collegeNodes) {
        colleges.push({ value: college.id, label: college.name });

        const majorNodes = Array.isArray(college.children) ? college.children : [];
        majors[college.id] = majorNodes.map((major: { id: string; name: string }) => ({
          value: major.id,
          label: major.name,
        }));

        for (const major of majorNodes) {
          const classNodes = Array.isArray(major.children) ? major.children : [];
          classes[major.id] = classNodes.map((item: { id: string; name: string }) => ({
            value: item.id,
            label: item.name,
          }));
        }
      }
    }

    const normalized: OptionsResponse = {
      colleges,
      majors,
      classes,
      grades: Array.isArray(gradesRes.data?.data?.items)
        ? gradesRes.data.data.items
        : [],
      user_statuses: [
        { value: "active", label: "启用" },
        { value: "inactive", label: "停用" },
        { value: "suspended", label: "冻结" },
      ],
      teacher_titles: Array.isArray(titlesRes.data?.data?.items)
        ? titlesRes.data.data.items
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
