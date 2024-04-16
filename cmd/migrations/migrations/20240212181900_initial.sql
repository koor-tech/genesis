-- +goose Up

create table
    public.providers (
                         id uuid not null default gen_random_uuid (),
                         name character varying null,
                         created_at timestamp with time zone not null default now(),
                         updated_at timestamp with time zone null default now(),
                         constraint providers_pkey primary key (id)
) tablespace pg_default;


create table
    public.customers (
                         id uuid not null default gen_random_uuid (),
                         company character varying not null,
                         email character varying not null,
                         created_at timestamp with time zone null default now(),
                         updated_at timestamp with time zone null default now(),
                         constraint customers_pkey primary key (id)
) tablespace pg_default;

create table
    public.clusters (
                        id uuid not null default gen_random_uuid (),
                        customer_id uuid null default gen_random_uuid (),
                        provider_id uuid null default gen_random_uuid (),
                        kube_config text null,
                        created_at timestamp with time zone null default now(),
                        updated_at timestamp with time zone null default now(),
                        constraint clusters_pkey primary key (id),
                        constraint public_clusters_customer_id_fkey foreign key (customer_id) references customers (id) on update cascade on delete cascade,
                        constraint public_clusters_provider_id_fkey foreign key (provider_id) references providers (id) on update cascade on delete cascade
) tablespace pg_default;

create table
    public.cluster_state (
                             id uuid not null default gen_random_uuid (),
                             phase smallint not null,
                             created_at timestamp with time zone null default now(),
                             updated_at timestamp with time zone null default now(),
                             cluster_id uuid null default gen_random_uuid (),
                             constraint cluster_state_pkey primary key (id),
                             constraint public_cluster_state_cluster_id_fkey foreign key (cluster_id) references clusters (id) on update cascade on delete cascade
) tablespace pg_default;

create table
    public.ssh (
                   id uuid not null default gen_random_uuid (),
                   cluster_id uuid not null,
                   private_file_path character varying not null,
                   public_file_path character varying not null,
                   private_key text not null,
                   public_key text not null,
                   constraint ssh_pkey primary key (id),
                   constraint public_ssh_cluster_id_fkey foreign key (cluster_id) references clusters (id) on update cascade on delete cascade
) tablespace pg_default;

-- +goose Down
DROP TABLE cluster_state;
DROP TABLE ssh;
DROP TABLE clusters;
DROP TABLE providers;
DROP TABLE customers;