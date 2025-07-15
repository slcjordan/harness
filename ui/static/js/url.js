function urlJoin(base, wsSlug, path) {
  let u = new URL([wsSlug, path].map(s => s.replace(/^\/+|\/+$/g, '')).join('/'), base);
  u.protocol = 'ws';
  return u.toString();
};
