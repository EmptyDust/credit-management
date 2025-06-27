import graphviz

# Create a new directed graph
dot = graphviz.Digraph('G', comment='Credit Management System Architecture')
dot.attr('graph',
         rankdir='TB',
         label='双创分申请平台 - 系统功能模块图 (System Function Module Diagram)',
         labelloc='t',
         fontsize='20',
         fontname='SimHei') # Use a font that supports Chinese characters

dot.attr('node', shape='box', style='rounded', fontname='SimHei')
dot.attr('edge', fontname='SimHei')

# Define Frontend Cluster
with dot.subgraph(name='cluster_frontend') as c:
    c.attr(label='前端 (Frontend)', style='filled', color='lightgrey', fontname='SimHei')
    c.node('frontend', '前端应用 (React + Vite)\n(UI & User Interaction)', shape='house', fillcolor='lightblue', style='filled')

# Define API Gateway
dot.node('gateway', 'API 网关 (API Gateway)\n- 路由转发\n- 统一认证', shape='Mdiamond', fillcolor='palegreen', style='filled')

# Define Backend Services Cluster
with dot.subgraph(name='cluster_backend') as c:
    c.attr(label='后端微服务 (Backend Microservices)', style='filled', color='lightyellow', fontname='SimHei')
    c.node('auth_service', '认证服务 (Auth Service)\n- 登录/Token验证\n- 权限管理', shape='component', fillcolor='lightpink', style='filled')
    c.node('user_service', '统一用户服务 (User Service)\n- 用户/学生/教师管理\n- 资料维护/搜索', shape='component', fillcolor='lightpink', style='filled')
    c.node('activity_service', '学分活动服务 (Credit Activity Service)\n- 活动/申请/参与者管理\n- 附件/统计', shape='component', fillcolor='lightpink', style='filled')

# Define Database Cluster
with dot.subgraph(name='cluster_db') as c:
    c.attr(label='数据存储 (Data Persistence)', style='filled', color='azure', fontname='SimHei')
    c.node('db', 'PostgreSQL 数据库', shape='cylinder', fillcolor='beige', style='filled')

# Define connections
dot.edge('frontend', 'gateway', label='HTTP/S 请求')
dot.edge('gateway', 'auth_service', label='认证/授权请求')
dot.edge('gateway', 'user_service', label='用户/学生/教师管理请求')
dot.edge('gateway', 'activity_service', label='活动/申请/附件请求')

dot.edge('auth_service', 'db', label='读/写用户信息')
dot.edge('user_service', 'db', label='读/写用户/学生/教师数据')
dot.edge('activity_service', 'db', label='读/写活动/申请/参与者数据')

# Render the graph to a file
# The output format is PNG. The filename will be 'system_function_module_diagram.gv'
# and the rendered image will be 'system_function_module_diagram.gv.png'.
dot.render('system_function_module_diagram', format='png', view=False, cleanup=True)

print("System function module diagram 'system_function_module_diagram.gv.png' has been generated.")