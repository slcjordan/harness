--
-- PostgreSQL database dump
--

-- Dumped from database version 16.9 (Debian 16.9-1.pgdg120+1)
-- Dumped by pg_dump version 16.9 (Debian 16.9-1.pgdg120+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: gitlab_user_cache; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.gitlab_user_cache (
    gitlab_user_id integer NOT NULL,
    email text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.gitlab_user_cache OWNER TO "user";

--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);


ALTER TABLE public.schema_migrations OWNER TO "user";

--
-- Name: slack_oauth_response; Type: TABLE; Schema: public; Owner: user
--

CREATE TABLE public.slack_oauth_response (
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    access_token text NOT NULL,
    app_id text NOT NULL,
    authed_user_access_token text NOT NULL,
    authed_user_expires_in integer NOT NULL,
    authed_user_id text NOT NULL,
    authed_user_refresh_token text NOT NULL,
    authed_user_scope text NOT NULL,
    authed_user_token_type text NOT NULL,
    bot_user_id text NOT NULL,
    enterprise_id text NOT NULL,
    enterprise_name text NOT NULL,
    error text NOT NULL,
    expires_in integer NOT NULL,
    incoming_webhook_channel text NOT NULL,
    incoming_webhook_channel_id text NOT NULL,
    incoming_webhook_configuration_url text NOT NULL,
    incoming_webhook_url text NOT NULL,
    is_enterprise_install boolean NOT NULL,
    metadata_cursor text NOT NULL,
    metadata_messages text[] NOT NULL,
    metadata_warnings text[] NOT NULL,
    ok boolean NOT NULL,
    refresh_token text NOT NULL,
    scope text NOT NULL,
    team_id text NOT NULL,
    team_name text NOT NULL,
    token_type text NOT NULL
);


ALTER TABLE public.slack_oauth_response OWNER TO "user";

--
-- Name: gitlab_user_cache gitlab_user_cache_pkey; Type: CONSTRAINT; Schema: public; Owner: user
--

ALTER TABLE ONLY public.gitlab_user_cache
    ADD CONSTRAINT gitlab_user_cache_pkey PRIMARY KEY (gitlab_user_id);


--
-- PostgreSQL database dump complete
--

