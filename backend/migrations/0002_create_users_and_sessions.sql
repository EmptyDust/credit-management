create table users (
    id varchar(26) primary key,
    username citext not null,
    password_hash text not null,
    email citext null,
    phone varchar(32) null,
    real_name varchar(128) not null,
    user_type varchar(16) not null,
    status varchar(16) not null,
    avatar_file_id varchar(26) null,
    department_id varchar(26) null,
    student_id varchar(64) null,
    teacher_id varchar(64) null,
    grade varchar(64) null,
    title varchar(128) null,
    refresh_token_version bigint not null default 1,
    last_login_at timestamptz null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz null
);

create unique index uq_users_username on users (username);
create unique index uq_users_student_id on users (student_id) where student_id is not null;
create unique index uq_users_teacher_id on users (teacher_id) where teacher_id is not null;
create unique index uq_users_email_active on users (email) where email is not null and deleted_at is null;
create unique index uq_users_phone_active on users (phone) where phone is not null and deleted_at is null;
create index idx_users_user_type_status on users (user_type, status) where deleted_at is null;
create index idx_users_department_id on users (department_id) where deleted_at is null;

create table auth_sessions (
    id varchar(26) primary key,
    user_id varchar(26) not null references users (id),
    refresh_token_hash text not null,
    issued_at timestamptz not null,
    expires_at timestamptz not null,
    revoked_at timestamptz null,
    revoke_reason text null,
    client_ip inet null,
    user_agent text null
);

create index idx_auth_sessions_user_id on auth_sessions (user_id);
create index idx_auth_sessions_expires_at on auth_sessions (expires_at);
