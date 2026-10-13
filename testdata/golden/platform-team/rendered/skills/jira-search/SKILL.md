---
name: jira-search
description: Search and summarize Jira tickets for platform-team, using the atlassian server.
---

Use the `atlassian` MCP server's Jira search tools to find tickets matching a query, then summarize the top results with ticket key, status, and assignee.

API-token sessions are not bound to a site, so pass the `cloudId` explicitly (call `getAccessibleAtlassianResources` once to find it). Jira Cloud rejects unbounded JQL such as a bare `ORDER BY created DESC`; always restrict the query, for example `project = KEY` or `created >= -30d`, and set `maxResults`.
