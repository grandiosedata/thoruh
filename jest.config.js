const configuration = {
  // clearMocks: true, // TODO: Figure out if I really want this enabled.
  collectCoverage: true,

  // globals: {},

  notify: true,
  // resetMocks: true, // TODO: Figure out if I really want this enabled.
  // resetModules: true, // TODO: Figure out if I really want this enabled.
  testEnvironment: 'node',

  testMatch: ['**/__tests__/**/*.js'],

  timers: 'fake',
  verbose: true
}

module.exports = configuration
