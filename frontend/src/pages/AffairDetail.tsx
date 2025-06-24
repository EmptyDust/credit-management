import { useEffect, useState } from "react";
import { useParams, Link, useNavigate } from "react-router-dom";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { 
  Users, 
  Award, 
  FileText, 
  ArrowLeft, 
  Download, 
  User, 
  Calendar, 
  CheckCircle, 
  XCircle, 
  Eye,
  Clock,
  TrendingUp,
  FileCheck,
  AlertCircle,
  Star
} from "lucide-react";
import apiClient from "@/lib/api";

interface Affair {
  id: number;
  name: string;
  description: string;
  category: string;
  status: string;
  max_credits: number;
  creator_id: string;
  attachments?: string;
  created_at?: string;
}

interface Participant {
  student_id: string;
  is_primary: boolean;
  role: string;
  name?: string;
}

interface AttachmentInfo {
  name: string;
  url: string;
}

interface Application {
  id: number;
  student_number: string;
  status: string;
  applied_credits: number;
  approved_credits: number;
  submission_time?: string;
}

export default function AffairDetailPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [affair, setAffair] = useState<Affair | null>(null);
  const [participants, setParticipants] = useState<Participant[]>([]);
  const [attachments, setAttachments] = useState<AttachmentInfo[]>([]);
  const [applications, setApplications] = useState<Application[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!id) return;
    setLoading(true);
    Promise.all([
      apiClient.get(`/affairs/${id}`),
      apiClient.get(`/affairs/${id}/participants`),
      apiClient.get(`/affairs/${id}/applications`)
    ]).then(([affairRes, partRes, appRes]) => {
      setAffair(affairRes.data.affair);
      setParticipants(partRes.data || []);
      setApplications(appRes.data.applications || []);
      try {
        setAttachments(affairRes.data.affair.attachments ? JSON.parse(affairRes.data.affair.attachments) : []);
      } catch {
        setAttachments([]);
      }
    }).finally(() => setLoading(false));
  }, [id]);

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="flex items-center gap-2">
          <Clock className="h-8 w-8 animate-spin" />
          <span className="text-lg">加载中...</span>
        </div>
      </div>
    );
  }
  if (!affair) {
    return (
      <div className="flex flex-col items-center mt-16">
        <AlertCircle className="h-16 w-16 text-red-500 mb-4" />
        <h2 className="text-xl font-semibold text-red-500 mb-2">未找到该事务</h2>
        <Button onClick={() => navigate('/affairs')}>返回事务列表</Button>
      </div>
    );
  }

  // 计算统计数据
  const approvedApps = applications.filter(app => app.status === 'approved');
  const pendingApps = applications.filter(app => app.status === 'pending');
  const rejectedApps = applications.filter(app => app.status === 'rejected');
  const approvalRate = applications.length > 0 ? Math.round((approvedApps.length / applications.length) * 100) : 0;
  const totalApprovedCredits = approvedApps.reduce((sum, app) => sum + (app.approved_credits || 0), 0);

  return (
    <div className="max-w-6xl mx-auto p-4 md:p-8 space-y-8">
      {/* 返回按钮 */}
      <Button variant="ghost" onClick={() => navigate(-1)} className="mb-2">
        <ArrowLeft className="h-4 w-4 mr-2" /> 返回
      </Button>

      {/* 事务基本信息 */}
      <Card className="rounded-xl shadow-lg bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-gray-900 dark:to-gray-800">
        <CardHeader>
          <div className="flex items-start justify-between">
            <div className="flex items-center gap-3">
              <div className="p-3 rounded-full bg-primary/10">
                <Award className="h-8 w-8 text-primary" />
              </div>
              <div>
                <CardTitle className="text-3xl font-bold">{affair.name}</CardTitle>
                <div className="flex items-center gap-2 mt-2">
                  <Badge className="bg-blue-100 text-blue-800 hover:bg-blue-200">{affair.category}</Badge>
                  <Badge className={affair.status === 'active' ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'}>
                    {affair.status === 'active' ? <CheckCircle className="w-3 h-3 mr-1" /> : <XCircle className="w-3 h-3 mr-1" />}
                    {affair.status === 'active' ? '活跃' : '停用'}
                  </Badge>
                </div>
              </div>
            </div>
            <div className="text-right">
              <div className="text-2xl font-bold text-primary">#{affair.id}</div>
              <div className="text-sm text-muted-foreground">事务ID</div>
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="text-lg leading-relaxed whitespace-pre-line">{affair.description}</div>
          
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="flex items-center gap-3 p-3 bg-white/50 dark:bg-gray-800/50 rounded-lg">
              <User className="h-5 w-5 text-blue-600" />
              <div>
                <div className="font-medium">创建人</div>
                <div className="text-sm text-muted-foreground">{affair.creator_id}</div>
              </div>
            </div>
            <div className="flex items-center gap-3 p-3 bg-white/50 dark:bg-gray-800/50 rounded-lg">
              <Star className="h-5 w-5 text-yellow-600" />
              <div>
                <div className="font-medium">最大学分</div>
                <div className="text-sm text-muted-foreground">{affair.max_credits} 学分</div>
              </div>
            </div>
            {affair.created_at && (
              <div className="flex items-center gap-3 p-3 bg-white/50 dark:bg-gray-800/50 rounded-lg">
                <Calendar className="h-5 w-5 text-green-600" />
                <div>
                  <div className="font-medium">创建时间</div>
                  <div className="text-sm text-muted-foreground">{new Date(affair.created_at).toLocaleDateString()}</div>
                </div>
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      {/* 统计卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card className="rounded-xl shadow-lg">
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-muted-foreground">参与学生</p>
                <p className="text-2xl font-bold">{participants.length}</p>
              </div>
              <Users className="h-8 w-8 text-blue-600" />
            </div>
          </CardContent>
        </Card>
        <Card className="rounded-xl shadow-lg">
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-muted-foreground">总申请数</p>
                <p className="text-2xl font-bold">{applications.length}</p>
              </div>
              <FileText className="h-8 w-8 text-green-600" />
            </div>
          </CardContent>
        </Card>
        <Card className="rounded-xl shadow-lg">
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-muted-foreground">通过率</p>
                <p className="text-2xl font-bold">{approvalRate}%</p>
              </div>
              <TrendingUp className="h-8 w-8 text-purple-600" />
            </div>
          </CardContent>
        </Card>
        <Card className="rounded-xl shadow-lg">
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-muted-foreground">已授学分</p>
                <p className="text-2xl font-bold">{totalApprovedCredits}</p>
              </div>
              <FileCheck className="h-8 w-8 text-orange-600" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* 参与学生 */}
      <Card className="rounded-xl shadow-lg">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Users className="h-5 w-5 text-blue-600" /> 参与学生
            <Badge variant="secondary">{participants.length} 人</Badge>
          </CardTitle>
        </CardHeader>
        <CardContent>
          {participants.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <Users className="h-12 w-12 mx-auto mb-2 opacity-50" />
              <p>暂无参与学生</p>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
              {participants.map((stu) => (
                <Link 
                  to={`/students?search=${stu.student_id}`} 
                  key={stu.student_id} 
                  className="flex items-center gap-3 p-3 bg-muted rounded-lg hover:bg-primary/10 transition-colors"
                >
                  <div className="p-2 rounded-full bg-primary/10">
                    <User className="h-4 w-4 text-primary" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="font-medium truncate">{stu.name || stu.student_id}</div>
                    <div className="text-sm text-muted-foreground">{stu.student_id}</div>
                  </div>
                  {stu.is_primary && (
                    <Badge className="ml-2" variant="secondary">
                      <Star className="w-3 h-3 mr-1" />
                      负责人
                    </Badge>
                  )}
                </Link>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* 附件列表 */}
      <Card className="rounded-xl shadow-lg">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FileText className="h-5 w-5 text-green-600" /> 附件
            <Badge variant="secondary">{attachments.length} 个</Badge>
          </CardTitle>
        </CardHeader>
        <CardContent>
          {attachments.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <FileText className="h-12 w-12 mx-auto mb-2 opacity-50" />
              <p>暂无附件</p>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
              {attachments.map((att, idx) => (
                <a 
                  key={att.url + idx} 
                  href={att.url} 
                  target="_blank" 
                  rel="noopener noreferrer" 
                  className="flex items-center gap-3 p-3 bg-muted rounded-lg hover:bg-primary/10 transition-colors"
                >
                  <div className="p-2 rounded-full bg-green-100">
                    <FileText className="h-4 w-4 text-green-600" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="font-medium truncate">{att.name}</div>
                    <div className="text-sm text-muted-foreground">点击下载</div>
                  </div>
                  <Download className="h-4 w-4 text-muted-foreground" />
                </a>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* 关联申请列表 */}
      <Card className="rounded-xl shadow-lg">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FileText className="h-5 w-5 text-purple-600" /> 关联申请
            <Badge variant="secondary">{applications.length} 个</Badge>
          </CardTitle>
        </CardHeader>
        <CardContent>
          {applications.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <FileText className="h-12 w-12 mx-auto mb-2 opacity-50" />
              <p>暂无关联申请</p>
            </div>
          ) : (
            <div className="space-y-4">
              {/* 申请状态统计 */}
              <div className="grid grid-cols-3 gap-4 mb-6">
                <div className="text-center p-3 bg-green-50 dark:bg-green-900/20 rounded-lg">
                  <div className="text-2xl font-bold text-green-600">{approvedApps.length}</div>
                  <div className="text-sm text-muted-foreground">已通过</div>
                </div>
                <div className="text-center p-3 bg-yellow-50 dark:bg-yellow-900/20 rounded-lg">
                  <div className="text-2xl font-bold text-yellow-600">{pendingApps.length}</div>
                  <div className="text-sm text-muted-foreground">待审核</div>
                </div>
                <div className="text-center p-3 bg-red-50 dark:bg-red-900/20 rounded-lg">
                  <div className="text-2xl font-bold text-red-600">{rejectedApps.length}</div>
                  <div className="text-sm text-muted-foreground">已拒绝</div>
                </div>
              </div>

              {/* 申请列表 */}
              <div className="overflow-x-auto">
                <table className="min-w-full text-sm">
                  <thead>
                    <tr className="bg-muted/60">
                      <th className="px-4 py-3 text-left font-medium">申请ID</th>
                      <th className="px-4 py-3 text-left font-medium">学生</th>
                      <th className="px-4 py-3 text-left font-medium">状态</th>
                      <th className="px-4 py-3 text-left font-medium">申请学分</th>
                      <th className="px-4 py-3 text-left font-medium">批准学分</th>
                      <th className="px-4 py-3 text-left font-medium">操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    {applications.map(app => (
                      <tr key={app.id} className="border-b hover:bg-muted/40 transition-colors">
                        <td className="px-4 py-3 font-medium">#{app.id}</td>
                        <td className="px-4 py-3">{app.student_number}</td>
                        <td className="px-4 py-3">
                          <Badge className={
                            app.status === 'approved' ? 'bg-green-100 text-green-800' : 
                            app.status === 'rejected' ? 'bg-red-100 text-red-800' : 
                            'bg-yellow-100 text-yellow-800'
                          }>
                            {app.status === 'approved' ? 'Approved' : app.status === 'rejected' ? 'Rejected' : app.status === 'pending' ? 'Pending' : 'Unsubmitted'}
                          </Badge>
                        </td>
                        <td className="px-4 py-3">{app.applied_credits}</td>
                        <td className="px-4 py-3">{app.approved_credits || '-'}</td>
                        <td className="px-4 py-3">
                          <Link 
                            to={`/applications?id=${app.id}`} 
                            className="inline-flex items-center gap-1 text-primary hover:underline"
                          >
                            <Eye className="h-4 w-4" />
                            查看详情
                          </Link>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
} 