import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useNavigate } from "react-router-dom";
import { Home, ArrowLeft, AlertTriangle } from "lucide-react";

export default function NotFound() {
  const navigate = useNavigate();

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-gray-50 to-blue-50 dark:from-gray-900 dark:to-gray-800 p-4">
      <Card className="w-full max-w-md text-center shadow-2xl border-0 bg-white/80 dark:bg-gray-900/80 backdrop-blur-sm">
        <CardHeader className="space-y-4">
          <div className="mx-auto w-20 h-20 bg-red-100 dark:bg-red-900/20 rounded-full flex items-center justify-center">
            <AlertTriangle className="h-10 w-10 text-red-600" />
          </div>
          <div>
            <CardTitle className="text-4xl font-bold text-gray-900 dark:text-gray-100">
              404
            </CardTitle>
            <CardDescription className="text-lg text-muted-foreground mt-2">
              页面未找到
            </CardDescription>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-muted-foreground">
            抱歉，您访问的页面不存在或已被移除。
          </p>
          <div className="flex flex-col sm:flex-row gap-3">
            <Button
              onClick={() => navigate(-1)}
              variant="outline"
              className="flex-1"
            >
              <ArrowLeft className="h-4 w-4 mr-2" />
              返回上页
            </Button>
            <Button onClick={() => navigate("/dashboard")} className="flex-1">
              <Home className="h-4 w-4 mr-2" />
              回到首页
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
