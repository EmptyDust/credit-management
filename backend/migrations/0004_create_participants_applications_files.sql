create table files (
    id varchar(26) primary key,
    storage_key text not null,
    original_name varchar(255) not null,
    mime_type varchar(255) not null,
    file_ext varchar(32) null,
    file_size bigint not null,
    md5_hash varchar(32) null,
    storage_backend varchar(32) not null,
    reference_count integer not null default 0,
    uploaded_by varchar(26) not null references users (id),
    uploaded_at timestamptz not null default now(),
    deleted_at timestamptz null
);

create unique index uq_files_storage_key on files (storage_key);
create index idx_files_uploaded_by on files (uploaded_by) where deleted_at is null;

alter table users
    add constraint fk_users_avatar_file
        foreign key (avatar_file_id) references files (id);

create table activity_files (
    id varchar(26) primary key,
    activity_id varchar(26) not null references activities (id),
    file_id varchar(26) not null references files (id),
    file_category varchar(64) null,
    description text null,
    download_count bigint not null default 0,
    created_at timestamptz not null default now(),
    deleted_at timestamptz null
);

create unique index uq_activity_files_active on activity_files (activity_id, file_id) where deleted_at is null;

create table avatar_files (
    id varchar(26) primary key,
    user_id varchar(26) not null references users (id),
    file_id varchar(26) not null references files (id),
    created_at timestamptz not null default now(),
    deleted_at timestamptz null
);

create unique index uq_avatar_files_user_active on avatar_files (user_id) where deleted_at is null;

create table activity_participants (
    id varchar(26) primary key,
    activity_id varchar(26) not null references activities (id),
    user_id varchar(26) not null references users (id),
    credits numeric(10,2) not null default 0 check (credits >= 0),
    joined_at timestamptz not null default now(),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz null
);

create unique index uq_activity_participants_active on activity_participants (activity_id, user_id) where deleted_at is null;
create index idx_activity_participants_user_id on activity_participants (user_id) where deleted_at is null;

create table applications (
    id varchar(26) primary key,
    activity_id varchar(26) not null references activities (id),
    user_id varchar(26) not null references users (id),
    status varchar(32) not null,
    applied_credits numeric(10,2) not null default 0 check (applied_credits >= 0),
    awarded_credits numeric(10,2) not null default 0 check (awarded_credits >= 0),
    submitted_at timestamptz not null default now(),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz null
);

create unique index uq_applications_active on applications (activity_id, user_id) where deleted_at is null;
create index idx_applications_user_status on applications (user_id, status) where deleted_at is null;
