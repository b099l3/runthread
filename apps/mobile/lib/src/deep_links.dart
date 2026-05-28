bool isRunthreadDeepLink(Uri uri) {
  return uri.scheme == 'runthread';
}

bool isStravaConnectionDeepLink(Uri uri) {
  return isRunthreadDeepLink(uri) &&
      uri.host == 'provider' &&
      uri.pathSegments.length >= 2 &&
      uri.pathSegments[0] == 'strava' &&
      const {'connected', 'callback', 'error'}.contains(uri.pathSegments[1]);
}
