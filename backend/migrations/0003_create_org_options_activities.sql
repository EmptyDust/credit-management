create table organization_nodes (
    id varchar(26) primary key,
    parent_id varchar(26) null references organization_nodes (id),
    node_type varchar(16) not null,
    name varchar(128) not null,
    code varchar(64) null,
    path text not null,
    level integer not null,
    sort_order integer not null default 0,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz null
);

create unique index uq_organization_nodes_path on organization_nodes (path) where deleted_at is null;

alter table users
    add constraint fk_users_department
        foreign key (department_id) references organization_nodes (id);

create table option_groups (
    id varchar(26) primary key,
    key varchar(64) not null,
    label varchar(128) not null,
    scope varchar(32) not null,
    version integer not null default 1,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz null
);

create unique index uq_option_groups_key_scope on option_groups (key, scope) where deleted_at is null;

create table option_items (
    id varchar(26) primary key,
    group_id varchar(26) not null references option_groups (id),
    value varchar(64) not null,
    label varchar(128) not null,
    sort_order integer not null default 0,
    enabled boolean not null default true,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz null
);

create unique index uq_option_items_group_value on option_items (group_id, value) where deleted_at is null;

create table activity_category_templates (
    id varchar(26) primary key,
    key varchar(64) not null,
    label varchar(128) not null,
    sort_order integer not null default 0,
    enabled boolean not null default true,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz null
);

create unique index uq_activity_category_templates_key on activity_category_templates (key) where deleted_at is null;

create table activity_field_schemas (
    id varchar(26) primary key,
    category_id varchar(26) not null references activity_category_templates (id),
    field_key varchar(64) not null,
    field_label varchar(128) not null,
    field_type varchar(32) not null,
    required boolean not null default false,
    default_value jsonb null,
    options_source varchar(128) null,
    validation_rules jsonb not null default '{}'::jsonb,
    display_order integer not null default 0,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz null
);

create unique index uq_activity_field_schemas_category_field on activity_field_schemas (category_id, field_key) where deleted_at is null;

create table activities (
    id varchar(26) primary key,
    title varchar(255) not null,
    description text null,
    category_id varchar(26) not null references activity_category_templates (id),
    academic_year varchar(32) not null,
    semester varchar(32) not null,
    start_date date not null,
    end_date date not null,
    status varchar(32) not null,
    owner_id varchar(26) not null references users (id),
    reviewer_id varchar(26) null references users (id),
    review_comments text null,
    reviewed_at timestamptz null,
    details_json jsonb not null default '{}'::jsonb,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    deleted_at timestamptz null
);

create index idx_activities_owner_status on activities (owner_id, status) where deleted_at is null;
create index idx_activities_category_status on activities (category_id, status) where deleted_at is null;
create index idx_activities_academic_period on activities (academic_year, semester) where deleted_at is null;
