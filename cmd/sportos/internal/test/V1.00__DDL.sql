-- user ddl
CREATE TABLE "user" (
	user_id character varying(40) not null,
    email character varying(200) not null,
    email_verified int not null,
    user_type character varying(40) not null,
    password_hash character varying(200) not null,
    token character varying(200),
    token_valid_until timestamp with time zone,
    token_refresh_until timestamp with time zone,
    created_at timestamp(6) with time zone not null,
    created_by character varying(40) not null,
    updated_at timestamp(6) with time zone,
    updated_by character varying(40),
    deleted_at timestamp(6) with time zone,
    deleted_by character varying(40),
	constraint pkuser PRIMARY KEY (user_id),
    constraint uniqueemail unique(email)
);

-- player ddl
CREATE TABLE player (
	user_id character varying(40) not null,
    name character varying(100) not null,
    city character varying(40) not null,
    preferences jsonb null,
    statistics jsonb null,
    created_at timestamp(6) with time zone not null,
    created_by character varying(40) not null,
    updated_at timestamp(6) with time zone,
    updated_by character varying(40),
    deleted_at timestamp(6) with time zone,
    deleted_by character varying(40),
	constraint pkplayer PRIMARY KEY (user_id),
    constraint fk_player_user_id foreign key (user_id)
    references "user" (user_id) match simple
);

-- coach ddl
CREATE TABLE coach (
	user_id character varying(40) not null,
    name character varying(100) not null,
    city character varying(40) not null,
    sport character varying(40) not null,
    booking jsonb null,
    reviews jsonb null,
    created_at timestamp(6) with time zone not null,
    created_by character varying(40) not null,
    updated_at timestamp(6) with time zone,
    updated_by character varying(40),
    deleted_at timestamp(6) with time zone,
    deleted_by character varying(40),
	constraint pkcoach PRIMARY KEY (user_id),
    constraint fk_coach_user_id foreign key (user_id)
    references "user" (user_id) match simple
);

-- place ddl
CREATE TABLE place (
	user_id character varying(40) not null,
    name character varying(100) not null,
    city character varying(40) not null,
    sport character varying(40) not null,
    booking jsonb null,
    reviews jsonb null,
    created_at timestamp(6) with time zone not null,
    created_by character varying(40) not null,
    updated_at timestamp(6) with time zone,
    updated_by character varying(40),
    deleted_at timestamp(6) with time zone,
    deleted_by character varying(40),
	constraint pkplace PRIMARY KEY (user_id),
    constraint fk_place_user_id foreign key (user_id)
    references "user" (user_id) match simple
);

create sequence event_id_seq
    start with 1000000000
    increment by 1
    no minvalue
    no maxvalue
    cache 1;

-- event ddl
CREATE TABLE event (
	event_id character varying(40) not null DEFAULT nextval('event_id_seq'::regclass),
    name character varying(100) not null,
    owner_id character varying(40) not null,
    sport character varying(40) not null,
    status character varying(40) not null,
    time timestamp(6) with time zone not null,
    teams jsonb null,
    tournament jsonb null,
    created_at timestamp(6) with time zone not null,
    created_by character varying(40) not null,
    updated_at timestamp(6) with time zone,
    updated_by character varying(40),
    deleted_at timestamp(6) with time zone,
    deleted_by character varying(40),
	constraint pkevent PRIMARY KEY (event_id),
    constraint fk_event_user_id foreign key (owner_id)
    references "place" (user_id) match simple
);

create sequence team_id_seq
    start with 1000000000
    increment by 1
    no minvalue
    no maxvalue
    cache 1;

-- team ddl
CREATE TABLE team (
	team_id character varying(40) not null DEFAULT nextval('team_id_seq'::regclass),
    name character varying(200) not null,
    sport character varying(40) not null,
    status character varying(40) not null,
    players character varying(1000) null,
    created_at timestamp(6) with time zone not null,
    created_by character varying(40) not null,
    updated_at timestamp(6) with time zone,
    updated_by character varying(40),
    deleted_at timestamp(6) with time zone,
    deleted_by character varying(40),
	constraint pkteam PRIMARY KEY (team_id)
);

create sequence match_id_seq
    start with 1000000000
    increment by 1
    no minvalue
    no maxvalue
    cache 1;

-- match ddl
CREATE TABLE match (
	match_id character varying(40) not null DEFAULT nextval('match_id_seq'::regclass), 
    status character varying(40) not null,
    start_time timestamp(6) with time zone not null,
    place_id character varying(200) not null,
    players character varying(1000) null,
    result character varying(20) null,
    sport character varying(40) not null,
    teams jsonb null,
    created_at timestamp(6) with time zone not null,
    created_by character varying(40) not null,
    updated_at timestamp(6) with time zone,
    updated_by character varying(40),
    deleted_at timestamp(6) with time zone,
    deleted_by character varying(40),
	constraint pkmatch PRIMARY KEY (match_id),
    constraint fk_match_place_id foreign key (place_id)
    references place (user_id) match simple
);

-- api_journal ddl
CREATE SEQUENCE api_journal_id_seq
    start with 1000000000
    increment by 1
    no minvalue
    no maxvalue
    cache 1;

CREATE TABLE api_journal (
  api_journal_id character varying(40) not null default nextval('api_journal_id_seq'::regclass),
  user_id character varying(40),
  request text,
  response text,
  request_json jsonb,
  response_json jsonb,
  source_ip text,
  created_at timestamp(6) with time zone not null,
  created_by character varying(40) not null,
  updated_at timestamp(6) with time zone,
  updated_by varchar(40) NULL,
  constraint pk_api_journal primary key (api_journal_id),
  constraint fk_aj_user foreign key (user_id)
  references "user" (user_id) match simple
);

comment on table api_journal is 'Contains all incoming requests';
comment on column api_journal.user_id is 'Id of user that initiated request';
comment on column api_journal.request is 'Full HTTP request text.';
comment on column api_journal.response is 'Full HTTP response text.';
comment on column api_journal.request_json is 'HTTP request json.';
comment on column api_journal.response_json is 'HTTP response json.';
comment on column api_journal.source_ip is 'IP adress of request.';

-- audit ddl
create sequence audit_id_seq
    start with 1000000000
    increment by 1
    no minvalue
    no maxvalue
    cache 1;

CREATE TABLE IF NOT EXISTS audit
(
    audit_id character varying(40) NOT NULL DEFAULT nextval('audit_id_seq'::regclass),
    entity character varying(100) not null,
    entity_id character varying(40),
    crud_action character varying(10),
    old jsonb,
    new jsonb,
    api_journal_id character varying(40),
    created_at timestamp(6) with time zone not null,
    created_by character varying(40) not null,
    constraint pk_audit PRIMARY KEY (audit_id),
    constraint fk_prj_api_journal foreign key (api_journal_id)
    references api_journal(api_journal_id) match simple
);

comment on table audit is 'Contains information about all entity changes.';
comment on column audit.entity is 'Which entity is being changed.';
comment on column audit.entity_id is 'ID of changed entity.';
comment on column audit.crud_action is 'Which crud action is being performed, e.g. C, U, D.';
comment on column audit.old is 'Old values of changed columns only.';
comment on column audit.new is 'New values of changed columns only.';
comment on column audit.api_journal_id is 'Id of api call that caused crud action.';

-- image sequence
create sequence image_seq start 1;

CREATE TABLE IF NOT EXISTS userpost (
    user_id text,
    user_text text,
    image_names text[],
    created_at timestamp(6) with time zone not null
);

comment on table userpost is 'Posts that users made.';
comment on column userpost.user_text is 'Text in player''s post';
comment on column userpost.image_names is 'Names of images in player''s post';

create sequence practice_id_seq
    start with 1000000000
    increment by 1
    no minvalue
    no maxvalue
    cache 1;

-- practice ddl
CREATE TABLE practice (
	practice_id character varying(40) not null DEFAULT nextval('practice_id_seq'::regclass), 
    player_id character varying(40),
    coach_id character varying(40),
    status character varying(40) not null,
    start_time timestamp(6) with time zone not null,
    sport character varying(40) not null,
    created_at timestamp(6) with time zone not null,
    created_by character varying(40) not null,
    updated_at timestamp(6) with time zone,
    updated_by character varying(40),
    deleted_at timestamp(6) with time zone,
    deleted_by character varying(40),
	constraint pkpractice PRIMARY KEY (practice_id),
    constraint fk_practice_player_id foreign key (player_id)
    references player (user_id) match simple,
    constraint fk_practice_coach_id foreign key (coach_id)
    references coach (user_id) match simple
);

create index place_reviews_index on place (reviews->'average');
create index coach_reviews_index on coach (reviews->'average');