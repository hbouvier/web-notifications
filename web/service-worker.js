'use strict';

const applicationServerPublicKey = '%{PUBLIC_KEY}%';

function urlB64ToUint8Array(base64String) {
  const padding = '='.repeat((4 - base64String.length % 4) % 4);
  const base64 = (base64String + padding)
    .replace(/\-/g, '+')
    .replace(/_/g, '/');

  const rawData = window.atob(base64);
  const outputArray = new Uint8Array(rawData.length);

  for (let i = 0; i < rawData.length; ++i) {
    outputArray[i] = rawData.charCodeAt(i);
  }
  return outputArray;
}

self.addEventListener('push', event => {
  let data = event.data.json()
  const title = data.title || 'Notification';
  const options = {
    body: data.body,
    icon: data.icon || 'images/icon.png',
    badge: data.badge || 'images/badge.png',
    data: data.data
  };
  event.waitUntil(self.registration.showNotification(title, options));
});

self.addEventListener('notificationclick', event => {
  let url = (event.notification.data || {href:'/'}).href;
  event.notification.close();
  event.waitUntil(
    clients.openWindow(url)
  );
});

self.addEventListener('pushsubscriptionchange', event => {
  console.log('pushsubscriptionchange: ', event);
  // TODO: Figure out how to pass the `email` herem to be able to 
  //       DELETE the old subscription and add the refreshed one.
  const applicationServerKey = urlB64ToUint8Array(applicationServerPublicKey);
  event.waitUntil(
    self.registration.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: applicationServerKey
    })
    .then(subscription => {
      // registerToServer(email, OLDsubscription, 'DELETE');
      // registerToServer(email, subscription, 'POST');
    })
  );
});

function registerToServer(email, subscription, verb) {
  let http = new XMLHttpRequest();
  let url = '/api/v1/register';
  let body = {
    subscriber: email,
    subscription: JSON.stringify(subscription)
  }
  http.open(verb, url, true);

  //Send the proper header information along with the request
  http.setRequestHeader("Content-type", "application/json; charset=UTF-8");

  http.onreadystatechange = function() {//Call a function when the state changes.
    if(http.readyState == 4 && http.status == 200) {
      console.log(http.responseText);
    }
  }
  return http.send(JSON.stringify(body));
}
