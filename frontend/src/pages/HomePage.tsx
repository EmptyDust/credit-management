import React, { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { Button } from '../components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';
import { Input } from '../components/ui/input';
import { Label } from '../components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../components/ui/select';
import { applicationAPI, affairAPI } from '../api';

interface Application {
  id: number;
  title: string;
  description: string;
  type: string;
  status: string;
  credits: number;
  createdAt: string;
}

interface Affair {
  id: number;
  title: string;
  description: string;
  type: string;
  status: string;
}

const HomePage: React.FC = () => {
  const { user, logout } = useAuth();
  const [applications, setApplications] = useState<Application[]>([]);
  const [affairs, setAffairs] = useState<Affair[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState('dashboard');

  // 申请表单状态
  const [applicationForm, setApplicationForm] = useState({
    title: '',
    description: '',
    type: '',
    affairId: '',
    credits: 0
  });

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      const [applicationsData, affairsData] = await Promise.all([
        applicationAPI.getApplications(),
        affairAPI.getAffairs()
      ]);
      setApplications(applicationsData);
      setAffairs(affairsData);
    } catch (error) {
      console.error('Failed to load data:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleApplicationSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await applicationAPI.createApplication({
        ...applicationForm,
        userId: user?.id,
        studentId: user?.id, // 简化处理
        affairId: parseInt(applicationForm.affairId)
      });
      setApplicationForm({
        title: '',
        description: '',
        type: '',
        affairId: '',
        credits: 0
      });
      loadData();
    } catch (error) {
      console.error('Failed to create application:', error);
    }
  };

  const handleLogout = () => {
    logout();
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'pending':
        return <span className="status-pending">待审核</span>;
      case 'approved':
        return <span className="status-approved">已通过</span>;
      case 'rejected':
        return <span className="status-rejected">已拒绝</span>;
      default:
        return <span className="bg-gray-100 text-gray-800 px-2 py-1 rounded-full text-xs font-medium">{status}</span>;
    }
  };

  const getTypeLabel = (type: string) => {
    const typeMap: { [key: string]: string } = {
      'innovation': '创新创业',
      'competition': '学科竞赛',
      'research': '科研项目',
      'internship': '实习实践',
      'other': '其他'
    };
    return typeMap[type] || type;
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 via-indigo-50 to-purple-50">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-lg text-gray-600">加载中...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 via-blue-50 to-indigo-50">
      {/* 导航栏 */}
      <nav className="bg-white/80 backdrop-blur-md shadow-sm border-b border-gray-200 sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex items-center">
              <div className="h-8 w-8 bg-gradient-to-r from-blue-600 to-purple-600 rounded-lg flex items-center justify-center mr-3">
                <span className="text-white text-sm font-bold">创</span>
              </div>
              <h1 className="text-xl font-semibold text-gray-900">
                创新创业学分管理平台
              </h1>
            </div>
            <div className="flex items-center space-x-4">
              <div className="flex items-center space-x-2">
                <div className="h-8 w-8 bg-gradient-to-r from-green-500 to-blue-500 rounded-full flex items-center justify-center">
                  <span className="text-white text-sm font-medium">
                    {user?.username?.charAt(0).toUpperCase()}
                  </span>
                </div>
                <div className="text-sm">
                  <p className="font-medium text-gray-900">{user?.username}</p>
                  <p className="text-gray-500 capitalize">{user?.role}</p>
                </div>
              </div>
              <Button
                variant="outline"
                onClick={handleLogout}
                className="border-gray-300 hover:bg-gray-50"
              >
                退出登录
              </Button>
            </div>
          </div>
        </div>
      </nav>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* 标签页导航 */}
        <div className="flex space-x-1 mb-8 bg-white rounded-lg p-1 shadow-sm">
          {[
            { id: 'dashboard', label: '仪表板', icon: '📊' },
            { id: 'applications', label: '我的申请', icon: '📝' },
            { id: 'new-application', label: '新建申请', icon: '➕' },
            ...(user?.role === 'admin' ? [{ id: 'admin', label: '管理面板', icon: '⚙️' }] : [])
          ].map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`flex-1 flex items-center justify-center px-4 py-2 rounded-md text-sm font-medium transition-all duration-200 ${activeTab === tab.id
                  ? 'bg-blue-600 text-white shadow-sm'
                  : 'text-gray-600 hover:text-gray-900 hover:bg-gray-50'
                }`}
            >
              <span className="mr-2">{tab.icon}</span>
              {tab.label}
            </button>
          ))}
        </div>

        {/* 仪表板 */}
        {activeTab === 'dashboard' && (
          <div className="space-y-6">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <Card className="card-shadow-hover border-0">
                <CardHeader className="pb-3">
                  <CardTitle className="text-lg flex items-center">
                    <span className="mr-2">📊</span>
                    总申请数
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="text-3xl font-bold text-blue-600">{applications.length}</div>
                  <p className="text-sm text-gray-500 mt-1">累计提交的申请数量</p>
                </CardContent>
              </Card>

              <Card className="card-shadow-hover border-0">
                <CardHeader className="pb-3">
                  <CardTitle className="text-lg flex items-center">
                    <span className="mr-2">🎓</span>
                    已获得学分
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="text-3xl font-bold text-green-600">
                    {applications
                      .filter(app => app.status === 'approved')
                      .reduce((sum, app) => sum + app.credits, 0)}
                  </div>
                  <p className="text-sm text-gray-500 mt-1">已通过申请的学分总和</p>
                </CardContent>
              </Card>

              <Card className="card-shadow-hover border-0">
                <CardHeader className="pb-3">
                  <CardTitle className="text-lg flex items-center">
                    <span className="mr-2">⏳</span>
                    待审核申请
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="text-3xl font-bold text-yellow-600">
                    {applications.filter(app => app.status === 'pending').length}
                  </div>
                  <p className="text-sm text-gray-500 mt-1">等待审核的申请数量</p>
                </CardContent>
              </Card>
            </div>

            {/* 最近申请 */}
            <Card className="card-shadow-hover border-0">
              <CardHeader>
                <CardTitle className="text-xl">最近申请</CardTitle>
                <CardDescription>查看您最近的学分申请记录</CardDescription>
              </CardHeader>
              <CardContent>
                {applications.length > 0 ? (
                  <div className="space-y-4">
                    {applications.slice(0, 5).map((application) => (
                      <div key={application.id} className="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
                        <div className="flex-1">
                          <h3 className="font-medium text-gray-900">{application.title}</h3>
                          <p className="text-sm text-gray-500 mt-1">{application.description}</p>
                          <div className="flex items-center space-x-2 mt-2">
                            <span className="text-xs bg-blue-100 text-blue-800 px-2 py-1 rounded">
                              {getTypeLabel(application.type)}
                            </span>
                            {getStatusBadge(application.status)}
                          </div>
                        </div>
                        <div className="text-right">
                          <div className="text-lg font-semibold text-gray-900">
                            {application.credits} 学分
                          </div>
                          <div className="text-xs text-gray-500">
                            {new Date(application.createdAt).toLocaleDateString()}
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="text-center py-8">
                    <div className="text-4xl mb-4">📝</div>
                    <p className="text-gray-500">暂无申请记录</p>
                    <Button
                      onClick={() => setActiveTab('new-application')}
                      className="mt-4 bg-blue-600 hover:bg-blue-700"
                    >
                      立即申请
                    </Button>
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        )}

        {/* 我的申请 */}
        {activeTab === 'applications' && (
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <h2 className="text-2xl font-bold text-gray-900">我的申请</h2>
              <Button
                onClick={() => setActiveTab('new-application')}
                className="bg-blue-600 hover:bg-blue-700"
              >
                新建申请
              </Button>
            </div>

            {applications.length > 0 ? (
              <div className="grid gap-6">
                {applications.map((application) => (
                  <Card key={application.id} className="card-shadow-hover border-0">
                    <CardHeader>
                      <div className="flex items-center justify-between">
                        <CardTitle className="text-lg">{application.title}</CardTitle>
                        {getStatusBadge(application.status)}
                      </div>
                      <CardDescription>{application.description}</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <div className="flex items-center justify-between">
                        <div className="flex items-center space-x-4">
                          <span className="text-sm bg-blue-100 text-blue-800 px-3 py-1 rounded-full">
                            {getTypeLabel(application.type)}
                          </span>
                          <span className="text-sm text-gray-500">
                            申请时间: {new Date(application.createdAt).toLocaleDateString()}
                          </span>
                        </div>
                        <div className="text-right">
                          <div className="text-2xl font-bold text-gray-900">
                            {application.credits} 学分
                          </div>
                          {application.status === 'approved' && (
                            <div className="text-sm text-green-600 font-medium">已认定</div>
                          )}
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                ))}
              </div>
            ) : (
              <Card className="card-shadow-hover border-0">
                <CardContent className="text-center py-12">
                  <div className="text-6xl mb-4">📝</div>
                  <h3 className="text-xl font-semibold text-gray-900 mb-2">暂无申请记录</h3>
                  <p className="text-gray-500 mb-6">开始您的第一个学分申请吧！</p>
                  <Button
                    onClick={() => setActiveTab('new-application')}
                    className="bg-blue-600 hover:bg-blue-700"
                  >
                    立即申请
                  </Button>
                </CardContent>
              </Card>
            )}
          </div>
        )}

        {/* 新建申请 */}
        {activeTab === 'new-application' && (
          <div className="max-w-2xl mx-auto">
            <div className="text-center mb-8">
              <h2 className="text-2xl font-bold text-gray-900 mb-2">新建申请</h2>
              <p className="text-gray-600">填写申请信息，提交您的学分申请</p>
            </div>

            <Card className="card-shadow-hover border-0">
              <CardContent className="pt-6">
                <form onSubmit={handleApplicationSubmit} className="space-y-6">
                  <div className="space-y-2">
                    <Label htmlFor="title" className="text-sm font-medium text-gray-700">
                      申请标题
                    </Label>
                    <Input
                      id="title"
                      value={applicationForm.title}
                      onChange={(e) => setApplicationForm(prev => ({ ...prev, title: e.target.value }))}
                      required
                      placeholder="请输入申请标题"
                      className="input-field"
                    />
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="description" className="text-sm font-medium text-gray-700">
                      申请描述
                    </Label>
                    <textarea
                      id="description"
                      value={applicationForm.description}
                      onChange={(e) => setApplicationForm(prev => ({ ...prev, description: e.target.value }))}
                      required
                      placeholder="请详细描述您的申请内容"
                      className="input-field min-h-[100px] resize-none"
                    />
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <Label htmlFor="type" className="text-sm font-medium text-gray-700">
                        申请类型
                      </Label>
                      <Select value={applicationForm.type} onValueChange={(value) => setApplicationForm(prev => ({ ...prev, type: value }))}>
                        <SelectTrigger className="input-field">
                          <SelectValue placeholder="请选择申请类型" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="innovation">创新创业</SelectItem>
                          <SelectItem value="competition">学科竞赛</SelectItem>
                          <SelectItem value="research">科研项目</SelectItem>
                          <SelectItem value="internship">实习实践</SelectItem>
                          <SelectItem value="other">其他</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="credits" className="text-sm font-medium text-gray-700">
                        申请学分
                      </Label>
                      <Input
                        id="credits"
                        type="number"
                        step="0.5"
                        value={applicationForm.credits}
                        onChange={(e) => setApplicationForm(prev => ({ ...prev, credits: parseFloat(e.target.value) || 0 }))}
                        required
                        placeholder="请输入申请学分"
                        className="input-field"
                      />
                    </div>
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="affairId" className="text-sm font-medium text-gray-700">
                      关联事项
                    </Label>
                    <Select value={applicationForm.affairId} onValueChange={(value) => setApplicationForm(prev => ({ ...prev, affairId: value }))}>
                      <SelectTrigger className="input-field">
                        <SelectValue placeholder="请选择关联事项" />
                      </SelectTrigger>
                      <SelectContent>
                        {affairs.map((affair) => (
                          <SelectItem key={affair.id} value={affair.id.toString()}>
                            {affair.title}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  <Button
                    type="submit"
                    className="w-full h-12 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700 text-white font-semibold rounded-lg transition-all duration-200 transform hover:scale-105"
                  >
                    提交申请
                  </Button>
                </form>
              </CardContent>
            </Card>
          </div>
        )}

        {/* 管理面板 */}
        {activeTab === 'admin' && user?.role === 'admin' && (
          <div className="space-y-6">
            <h2 className="text-2xl font-bold text-gray-900">管理面板</h2>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <Card className="card-shadow-hover border-0">
                <CardHeader>
                  <CardTitle className="text-lg flex items-center">
                    <span className="mr-2">📋</span>
                    事项管理
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-gray-600 mb-4">管理创新创业事项和活动</p>
                  <Button className="w-full bg-blue-600 hover:bg-blue-700">
                    查看事项
                  </Button>
                </CardContent>
              </Card>

              <Card className="card-shadow-hover border-0">
                <CardHeader>
                  <CardTitle className="text-lg flex items-center">
                    <span className="mr-2">✅</span>
                    申请审核
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-gray-600 mb-4">审核学生的学分申请</p>
                  <Button className="w-full bg-green-600 hover:bg-green-700">
                    审核申请
                  </Button>
                </CardContent>
              </Card>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default HomePage; 