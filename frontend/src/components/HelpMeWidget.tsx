import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import {
  HelpCircle,
  Eye,
  Edit,
  Trash,
  PlusCircle,
  RefreshCw,
  Search,
  Upload,
  Download,
  Clock,
  AlertCircle,
  CheckCircle,
  XCircle,
  FileText,
  MousePointer2,
  GitBranch,
  MessageCircle,
  Sparkles,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";

export function HelpMeWidget() {
  return (
    <Sheet>
      <SheetTrigger asChild>
        <button
          className="fixed bottom-6 right-6 p-4 bg-gradient-to-br from-primary to-primary/80 text-primary-foreground rounded-full shadow-lg hover:shadow-xl hover:scale-110 transition-all duration-300 z-50 group"
          aria-label="打开帮助面板"
        >
          <HelpCircle className="w-6 h-6 group-hover:rotate-12 transition-transform" />
          <span className="absolute -top-1 -right-1 flex h-3 w-3">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-primary/60 opacity-75"></span>
            <span className="relative inline-flex rounded-full h-3 w-3 bg-primary-foreground/90"></span>
          </span>
        </button>
      </SheetTrigger>
      <SheetContent className="overflow-y-auto w-full sm:max-w-lg">
        <SheetHeader className="pb-4 border-b">
          <SheetTitle className="flex items-center gap-2 text-xl">
            <Sparkles className="w-5 h-5 text-primary" />
            操作指引 & 帮助
          </SheetTitle>
          <SheetDescription>
            遇到问题？这里有你需要的所有解释。
          </SheetDescription>
        </SheetHeader>

        <div className="mt-6 space-y-8">
          {/* 图标说明 */}
          <section>
            <h3 className="font-bold text-lg mb-3 flex items-center gap-2 text-foreground">
              <MousePointer2 className="w-5 h-5 text-primary" />
              图标说明
            </h3>
            <div className="space-y-3 text-sm bg-muted/50 p-4 rounded-lg">
              <div className="flex items-start gap-3">
                <div className="p-1.5 rounded-md bg-blue-100 dark:bg-blue-900/50">
                  <Eye className="w-4 h-4 text-blue-600 dark:text-blue-400" />
                </div>
                <div>
                  <span className="font-medium">查看详情</span>
                  <p className="text-muted-foreground text-xs mt-0.5">
                    点击进入详情页，查看活动的完整描述、附件和当前审核进度。
                  </p>
                </div>
              </div>
              <div className="flex items-start gap-3">
                <div className="p-1.5 rounded-md bg-amber-100 dark:bg-amber-900/50">
                  <Edit className="w-4 h-4 text-amber-600 dark:text-amber-400" />
                </div>
                <div>
                  <span className="font-medium">编辑</span>
                  <p className="text-muted-foreground text-xs mt-0.5">
                    <strong>重点</strong>：只有在"草稿"或"被驳回"状态下，且你是活动的创建者时，这个按钮才会出现。
                  </p>
                </div>
              </div>
              <div className="flex items-start gap-3">
                <div className="p-1.5 rounded-md bg-red-100 dark:bg-red-900/50">
                  <Trash className="w-4 h-4 text-red-600 dark:text-red-400" />
                </div>
                <div>
                  <span className="font-medium">删除</span>
                  <p className="text-muted-foreground text-xs mt-0.5">
                    永久删除该记录。<strong className="text-red-500">警告</strong>：已通过审核的活动通常无法删除。
                  </p>
                </div>
              </div>
              <div className="flex items-start gap-3">
                <div className="p-1.5 rounded-md bg-green-100 dark:bg-green-900/50">
                  <PlusCircle className="w-4 h-4 text-green-600 dark:text-green-400" />
                </div>
                <div>
                  <span className="font-medium">新建活动</span>
                  <p className="text-muted-foreground text-xs mt-0.5">
                    发起一个新的学分申请录入。
                  </p>
                </div>
              </div>
              <div className="flex items-start gap-3">
                <div className="p-1.5 rounded-md bg-gray-100 dark:bg-gray-800">
                  <RefreshCw className="w-4 h-4 text-gray-600 dark:text-gray-400" />
                </div>
                <div>
                  <span className="font-medium">刷新/重置</span>
                  <p className="text-muted-foreground text-xs mt-0.5">
                    清空所有搜索和筛选条件，重新加载列表。
                  </p>
                </div>
              </div>
              <div className="flex items-start gap-3">
                <div className="p-1.5 rounded-md bg-purple-100 dark:bg-purple-900/50">
                  <Upload className="w-4 h-4 text-purple-600 dark:text-purple-400" />
                </div>
                <div>
                  <span className="font-medium">导入</span>
                  <p className="text-muted-foreground text-xs mt-0.5">
                    批量导入活动数据（仅支持 .xlsx 或 .csv 格式）。
                  </p>
                </div>
              </div>
              <div className="flex items-start gap-3">
                <div className="p-1.5 rounded-md bg-teal-100 dark:bg-teal-900/50">
                  <Download className="w-4 h-4 text-teal-600 dark:text-teal-400" />
                </div>
                <div>
                  <span className="font-medium">导出</span>
                  <p className="text-muted-foreground text-xs mt-0.5">
                    将当前列表数据导出为 Excel 文件。
                  </p>
                </div>
              </div>
            </div>
          </section>

          {/* 状态徽章 */}
          <section>
            <h3 className="font-bold text-lg mb-3 flex items-center gap-2 text-foreground">
              <GitBranch className="w-5 h-5 text-primary" />
              状态流转
            </h3>
            <div className="space-y-3 text-sm">
              <div className="flex items-start gap-3 p-3 rounded-lg border bg-card">
                <Badge
                  variant="default"
                  className="bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200 shrink-0"
                >
                  <Clock className="w-3 h-3 mr-1" />
                  草稿
                </Badge>
                <p className="text-muted-foreground">
                  仅你自己可见，尚未提交给老师，随时可修改。记得提交哦！
                </p>
              </div>
              <div className="flex items-start gap-3 p-3 rounded-lg border bg-card">
                <Badge
                  variant="default"
                  className="bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200 shrink-0"
                >
                  <AlertCircle className="w-3 h-3 mr-1" />
                  待审核
                </Badge>
                <p className="text-muted-foreground">
                  已提交，正在等待老师审批。此时<strong>不可编辑</strong>。
                </p>
              </div>
              <div className="flex items-start gap-3 p-3 rounded-lg border bg-card">
                <Badge
                  variant="default"
                  className="bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200 shrink-0"
                >
                  <CheckCircle className="w-3 h-3 mr-1" />
                  已通过
                </Badge>
                <p className="text-muted-foreground">
                  恭喜！老师已确认，学分已生效。<strong>不可修改或删除</strong>。
                </p>
              </div>
              <div className="flex items-start gap-3 p-3 rounded-lg border bg-card">
                <Badge
                  variant="default"
                  className="bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200 shrink-0"
                >
                  <XCircle className="w-3 h-3 mr-1" />
                  已驳回
                </Badge>
                <p className="text-muted-foreground">
                  申请被拒绝，请查看详情里的"审核意见"，修改后可再次提交。
                </p>
              </div>
            </div>
          </section>

          {/* 交互组件指引 */}
          <section>
            <h3 className="font-bold text-lg mb-3 flex items-center gap-2 text-foreground">
              <Search className="w-5 h-5 text-primary" />
              交互组件操作指引
            </h3>
            <div className="space-y-4 text-sm">
              <div className="p-4 rounded-lg border-l-4 border-l-blue-500 bg-blue-50/50 dark:bg-blue-950/20">
                <h4 className="font-semibold text-blue-700 dark:text-blue-400 mb-1">
                  搜索与筛选
                </h4>
                <ul className="list-disc pl-4 space-y-1 text-muted-foreground">
                  <li>输入关键词后请按 <kbd className="px-1.5 py-0.5 text-xs bg-muted rounded font-mono">Enter</kbd> 或点击放大镜图标进行搜索</li>
                  <li>分类筛选器是<strong>实时生效</strong>的，选择后立即过滤</li>
                  <li>点击"重置"按钮可清空所有筛选条件</li>
                </ul>
              </div>
              <div className="p-4 rounded-lg border-l-4 border-l-purple-500 bg-purple-50/50 dark:bg-purple-950/20">
                <h4 className="font-semibold text-purple-700 dark:text-purple-400 mb-1">
                  批量导入
                </h4>
                <ul className="list-disc pl-4 space-y-1 text-muted-foreground">
                  <li>仅支持 <code className="px-1 py-0.5 text-xs bg-muted rounded">.xlsx</code> 或 <code className="px-1 py-0.5 text-xs bg-muted rounded">.csv</code> 格式</li>
                  <li><strong>强烈建议</strong>先下载模板，按模板格式填写</li>
                  <li>请勿修改模板的表头，否则可能导入失败</li>
                </ul>
              </div>
              <div className="p-4 rounded-lg border-l-4 border-l-amber-500 bg-amber-50/50 dark:bg-amber-950/20">
                <h4 className="font-semibold text-amber-700 dark:text-amber-400 mb-1">
                  表单验证
                </h4>
                <ul className="list-disc pl-4 space-y-1 text-muted-foreground">
                  <li>标有 <span className="text-red-500">*</span> 号的字段为必填项</li>
                  <li>如果提交按钮不可点，请检查是否有必填项未填</li>
                  <li>学分不能为负数，日期需填写有效格式</li>
                </ul>
              </div>
            </div>
          </section>

          {/* 业务流程解惑 */}
          <section>
            <h3 className="font-bold text-lg mb-3 flex items-center gap-2 text-foreground">
              <FileText className="w-5 h-5 text-primary" />
              常见场景指引
            </h3>
            <div className="space-y-4 text-sm">
              <div className="p-4 rounded-lg bg-gradient-to-r from-green-50 to-emerald-50 dark:from-green-950/30 dark:to-emerald-950/30 border">
                <h4 className="font-semibold text-green-700 dark:text-green-400 mb-2 flex items-center gap-1">
                  <span className="w-5 h-5 rounded-full bg-green-500 text-white text-xs flex items-center justify-center">1</span>
                  我参加了一个比赛，怎么录入？
                </h4>
                <ol className="list-decimal pl-5 space-y-1 text-muted-foreground">
                  <li>点击"新建活动"按钮</li>
                  <li>填写比赛名称、时间、类别等信息</li>
                  <li><strong>关键</strong>：在附件处上传获奖证书（支持 JPG/PNG/PDF）</li>
                  <li>点击"提交审核"（保存草稿不会提交给老师）</li>
                </ol>
              </div>
              <div className="p-4 rounded-lg bg-gradient-to-r from-amber-50 to-orange-50 dark:from-amber-950/30 dark:to-orange-950/30 border">
                <h4 className="font-semibold text-amber-700 dark:text-amber-400 mb-2 flex items-center gap-1">
                  <span className="w-5 h-5 rounded-full bg-amber-500 text-white text-xs flex items-center justify-center">2</span>
                  我发现提交的信息填错了，怎么办？
                </h4>
                <ul className="list-disc pl-5 space-y-1 text-muted-foreground">
                  <li>如果状态是<Badge variant="outline" className="mx-1 text-xs">待审核</Badge>：需要先联系老师驳回，或点击"撤回"按钮（如有）</li>
                  <li>如果状态是<Badge variant="outline" className="mx-1 text-xs">草稿</Badge>：直接点击列表右侧的"编辑"图标</li>
                </ul>
              </div>
              <div className="p-4 rounded-lg bg-gradient-to-r from-blue-50 to-indigo-50 dark:from-blue-950/30 dark:to-indigo-950/30 border">
                <h4 className="font-semibold text-blue-700 dark:text-blue-400 mb-2 flex items-center gap-1">
                  <span className="w-5 h-5 rounded-full bg-blue-500 text-white text-xs flex items-center justify-center">3</span>
                  为什么我看不到"编辑"按钮？
                </h4>
                <p className="text-muted-foreground pl-5">
                  系统权限控制：你只能编辑<strong>你自己创建的</strong>且<strong>未归档/未通过</strong>的活动。
                  如果活动已通过，需联系管理员修改。
                </p>
              </div>
            </div>
          </section>

          {/* FAQ */}
          <section>
            <h3 className="font-bold text-lg mb-3 flex items-center gap-2 text-foreground">
              <MessageCircle className="w-5 h-5 text-primary" />
              常见问题 FAQ
            </h3>
            <div className="space-y-3 text-sm">
              <details className="group p-3 rounded-lg border bg-card hover:bg-muted/50 transition-colors cursor-pointer">
                <summary className="font-medium list-none flex items-center justify-between">
                  <span>页面一直转圈加载不出来？</span>
                  <span className="text-muted-foreground group-open:rotate-180 transition-transform">▼</span>
                </summary>
                <p className="mt-2 text-muted-foreground pl-2 border-l-2 border-primary/30">
                  请检查网络连接，或尝试点击右上角的"刷新"按钮。如果顶部进度条走完仍无数据，可能是服务器正在维护。
                </p>
              </details>
              <details className="group p-3 rounded-lg border bg-card hover:bg-muted/50 transition-colors cursor-pointer">
                <summary className="font-medium list-none flex items-center justify-between">
                  <span>手机上表格显示不全？</span>
                  <span className="text-muted-foreground group-open:rotate-180 transition-transform">▼</span>
                </summary>
                <p className="mt-2 text-muted-foreground pl-2 border-l-2 border-primary/30">
                  系统支持响应式布局，但在手机端表格会出现横向滚动条，请<strong>左右滑动表格区域</strong>查看操作按钮。
                </p>
              </details>
              <details className="group p-3 rounded-lg border bg-card hover:bg-muted/50 transition-colors cursor-pointer">
                <summary className="font-medium list-none flex items-center justify-between">
                  <span>上传图片/附件失败？</span>
                  <span className="text-muted-foreground group-open:rotate-180 transition-transform">▼</span>
                </summary>
                <p className="mt-2 text-muted-foreground pl-2 border-l-2 border-primary/30">
                  请检查文件大小是否超过系统限制（建议不超过 5MB），可尝试使用压缩后的图片或 PDF 格式。
                </p>
              </details>
              <details className="group p-3 rounded-lg border bg-card hover:bg-muted/50 transition-colors cursor-pointer">
                <summary className="font-medium list-none flex items-center justify-between">
                  <span>导入数据报错？</span>
                  <span className="text-muted-foreground group-open:rotate-180 transition-transform">▼</span>
                </summary>
                <p className="mt-2 text-muted-foreground pl-2 border-l-2 border-primary/30">
                  请确保：1) 文件格式为 .xlsx 或 .csv；2) 使用官方提供的模板；3) 表头未被修改；4) 数据格式正确（日期、数字等）。
                </p>
              </details>
              <details className="group p-3 rounded-lg border bg-card hover:bg-muted/50 transition-colors cursor-pointer">
                <summary className="font-medium list-none flex items-center justify-between">
                  <span>如何联系管理员？</span>
                  <span className="text-muted-foreground group-open:rotate-180 transition-transform">▼</span>
                </summary>
                <p className="mt-2 text-muted-foreground pl-2 border-l-2 border-primary/30">
                  如遇到无法自行解决的问题，请联系你的辅导员或班主任，或发送邮件至学院教务处。
                </p>
              </details>
            </div>
          </section>

          {/* 底部提示 */}
          <div className="pt-4 border-t text-center text-xs text-muted-foreground">
            <p>💡 小提示：本帮助面板可随时点击右下角按钮打开</p>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  );
}

