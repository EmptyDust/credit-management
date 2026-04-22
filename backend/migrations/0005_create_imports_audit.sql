create table import_jobs (
    id varchar(26) primary key,
    job_type varchar(32) not null,
    initiator_id varchar(26) not null references users (id),
    file_name varchar(255) not null,
    file_hash varchar(128) not null,
    stage varchar(32) not null,
    expires_at timestamptz not null,
    summary_json jsonb not null default '{}'::jsonb,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create index idx_import_jobs_initiator_stage on import_jobs (initiator_id, stage);
create index idx_import_jobs_expires_at on import_jobs (expires_at);

create table import_job_rows (
    id varchar(26) primary key,
    job_id varchar(26) not null references import_jobs (id) on delete cascade,
    row_no integer not null,
    payload_json jsonb not null,
    validation_status varchar(32) not null,
    validation_errors jsonb not null default '[]'::jsonb,
    created_at timestamptz not null default now()
);

create unique index uq_import_job_rows_job_row on import_job_rows (job_id, row_no);

create table audit_logs (
    id varchar(26) primary key,
    actor_id varchar(26) null references users (id),
    actor_role varchar(16) not null,
    target_type varchar(64) not null,
    target_id varchar(26) null,
    action varchar(64) not null,
    result varchar(32) not null,
    source_ip inet null,
    request_id varchar(64) null,
    summary_change jsonb not null default '{}'::jsonb,
    created_at timestamptz not null default now()
);

create index idx_audit_logs_actor_id on audit_logs (actor_id);
create index idx_audit_logs_target on audit_logs (target_type, target_id);
create index idx_audit_logs_created_at on audit_logs (created_at);
