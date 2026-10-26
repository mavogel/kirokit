import test from 'node:test';
import assert from 'node:assert/strict';
import { feedbackURL } from './feedback.js';

const parse = (url) => {
  const u = new URL(url);
  return { origin: u.origin, path: u.pathname, title: u.searchParams.get('title'), body: u.searchParams.get('body'), labels: u.searchParams.get('labels') };
};

test('Feedback opens the new-issue form of the team repository, prefilled for the owner', () => {
  const got = parse(feedbackURL('acme/kiro-config', 'product-team', 'agent', 'product-owner', '@nick'));

  assert.equal(got.origin, 'https://github.com');
  assert.equal(got.path, '/acme/kiro-config/issues/new');
  assert.equal(got.title, 'Feedback: product-team agent product-owner');
  assert.equal(got.labels, 'feedback,team:product-team,agent:product-owner');
  assert.match(got.body, /^\*\*Team:\*\* product-team\n\*\*Agent:\*\* product-owner\n\*\*Owner:\*\* @nick\n\n\*\*What happened\*\*\n/);
  assert.match(got.body, /\*\*What you expected\*\*\n/);
  assert.match(got.body, /_Sent from the kirokit desktop app\._$/);
});

test('A skill is labelled as a skill, and an item without owner mentions nobody', () => {
  const got = parse(feedbackURL('acme/cfg', 'alpha', 'skill', 'jira-search', ''));

  assert.equal(got.title, 'Feedback: alpha skill jira-search');
  assert.equal(got.labels, 'feedback,team:alpha,skill:jira-search');
  assert.match(got.body, /\*\*Skill:\*\* jira-search\n/);
  assert.doesNotMatch(got.body, /Owner|@/);
});

// Wails' BrowserOpenURL (v2.17, internal/frontend/utils/urlValidator.go)
// refuses an address holding any of these characters, so the address must
// percent-encode them, including the ones encodeURIComponent leaves alone.
const WAILS_REFUSES = /[;|`$\\<>*{}[\]()~! \t\n\r]/;

test('The address passes the browser opener: no shell metacharacters survive encoding', () => {
  const url = feedbackURL('acme/cfg', 'te(am)', 'agent', "a*b!c~d'e (x) [y] {z} <w> $v;u|t`s\\r", '@o');
  assert.doesNotMatch(url, WAILS_REFUSES);
  const got = parse(url);
  assert.equal(got.title, "Feedback: te(am) agent a*b!c~d'e (x) [y] {z} <w> $v;u|t`s\\r");
  assert.match(got.body, /^\*\*Team:\*\* te\(am\)\n/);
});

test('Names with characters that mean something in an address arrive unchanged', () => {
  const name = 'a&b #c "d" e+f?g=h%';
  const got = parse(feedbackURL('acme/cfg', 'te am&x', 'agent', name, '@o'));

  assert.equal(got.path, '/acme/cfg/issues/new');
  assert.equal(got.title, `Feedback: te am&x agent ${name}`);
  assert.equal(got.labels, `feedback,team:te am&x,agent:${name}`);
  assert.match(got.body, /\*\*Agent:\*\* a&b #c "d" e\+f\?g=h%\n/);
});
