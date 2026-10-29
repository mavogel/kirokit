import test from 'node:test';
import assert from 'node:assert/strict';
import { feedbackURL, onFeedbackClick } from './feedback.js';

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

// Wails v3's Browser.OpenURL (pkg/application/urlvalidator.go) refuses an
// address holding ; | ` $ \ < > or whitespace; Wails v2.17 also refused
// * { } [ ] ( ) ~ !. The address percent-encodes all of them, including the ones
// encodeURIComponent leaves alone, so issue addresses stay as they were.
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

// The button's data attributes become the item the issue is about: kind from
// data-feedback, team and name from theirs. A click inside the button (on its
// text) counts; a disabled button and a click elsewhere do nothing.
test('A Feedback press reports the pressed item, and only for an enabled Feedback button', () => {
  let handler;
  const section = { addEventListener: (type, fn) => { assert.equal(type, 'click'); handler = fn; } };
  const pressed = [];
  onFeedbackClick(section, (item) => pressed.push(item));

  const button = (disabled) => ({ disabled, dataset: { feedback: 'skill', team: 'alpha', name: 'jira-search' } });
  const target = (found) => ({ closest: (sel) => (sel === 'button[data-feedback]' ? found : null) });

  handler({ target: target(button(false)) });
  handler({ target: target(button(true)) });
  handler({ target: target(null) });

  assert.deepEqual(pressed, [{ team: 'alpha', kind: 'skill', name: 'jira-search' }]);
});
