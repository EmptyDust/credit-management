import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

export default function ApiTesterPage() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>API 测试器</CardTitle>
      </CardHeader>
      <CardContent className="text-sm text-muted-foreground">
        首期重构中不开放浏览器内 API 测试器执行能力。该入口保留为后续非生产调试模式的占位页。
      </CardContent>
    </Card>
  );
}
