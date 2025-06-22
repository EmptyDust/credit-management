import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { jwtDecode } from 'jwt-decode';
import { Button } from "@/components/ui/button";

interface UserPayload {
  username: string;
  role: string;
  exp: number;
}

export default function HomePage() {
  const navigate = useNavigate();
  const token = localStorage.getItem('token');

  if (!token) {
    // 理论上ProtectedRoute会处理，但作为双重保障
    navigate('/login');
    return null;
  }

  const user: UserPayload = jwtDecode(token);

  const handleLogout = () => {
    localStorage.removeItem('token');
    navigate('/login');
  };

  return (
    <div className="container p-4 mx-auto">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">创新创业学分管理平台</h1>
        <Button onClick={handleLogout} variant="outline">登出</Button>
      </div>

      <div className="p-8 mt-10 text-center bg-white rounded-lg shadow-md">
        <h2 className="text-2xl font-bold text-gray-800">欢迎, {user.username}!</h2>
        <p className="mt-2 text-lg text-gray-600">您的角色是: <span className="font-semibold capitalize">{user.role === 'student' ? '学生' : '管理员'}</span></p>
      </div>

      {/* 后续将在这里添加学分申请列表等内容 */}
    </div>
  );
} 