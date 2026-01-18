export const environment = {
  production: true,
  apiUrl: '/api',
  wsUrl: `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/ws`,
  appName: 'Waiting Room System'
};
