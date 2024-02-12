-- +goose Up
CREATE TABLE clients (
    id UUID PRIMARY KEY,
    name VARCHAR(255),
    created_at TIMESTAMP DEFAULT now() NOT NULL,
    updated_at TIMESTAMP DEFAULT now() NOT NULL
);

CREATE TABLE providers (
    id UUID PRIMARY KEY,
    name VARCHAR(255),
    created_at TIMESTAMP DEFAULT now() NOT NULL,
    updated_at TIMESTAMP DEFAULT now() NOT NULL
);

CREATE TABLE clusters (
   id UUID PRIMARY KEY,
   client_id UUID,
   provider_id UUID,
   created_at TIMESTAMP DEFAULT now() NOT NULL,
   updated_at TIMESTAMP DEFAULT now() NOT NULL
);

 ALTER TABLE clusters
   ADD CONSTRAINT fk_clusters_client_id
   FOREIGN KEY (client_id)
   REFERENCES clients(id)
   ON DELETE CASCADE;

 ALTER TABLE clusters
   ADD CONSTRAINT fk_clusters_provider_id
   FOREIGN KEY (provider_id)
   REFERENCES providers(id)
   ON DELETE CASCADE;

CREATE TABLE cluster_state (
   id UUID PRIMARY KEY,
   cluster_id UUID,
   phase integer,
   time TIMESTAMP
);

 ALTER TABLE cluster_state
   ADD CONSTRAINT fk_cluster_state_cluster_id
   FOREIGN KEY (cluster_id)
   REFERENCES clusters(id)
   ON DELETE CASCADE;

-- +goose Down
DROP TABLE cluster_state;
DROP TABLE clusters;
DROP TABLE providers;
DROP TABLE clients;