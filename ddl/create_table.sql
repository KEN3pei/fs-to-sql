CREATE TABLE public.directories (
	"path" varchar NOT NULL,
	"perm" varchar NOT NULL,
	"parent_path" varchar NOT NULL,
	"owner" varchar NOT NULL,
	"group" varchar NOT NULL,
	"size" integer NOT NULL,
	"updated_at" timestamp NOT NULL,
    "name" varchar NOT NULL
);

CREATE TABLE public.files (
	"path" varchar NOT NULL,
	"perm" varchar NOT NULL,
	"parent_path" varchar NOT NULL,
	"owner" varchar NOT NULL,
	"group" varchar NOT NULL,
	"size" integer NOT NULL,
	"updated_at" timestamp NOT NULL,
	"name" varchar NOT NULL
);
