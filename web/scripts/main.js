'use strict';
const applicationServerPublicKey = '%{PUBLIC_KEY}%';
const primaryButton = document.querySelector('.btn-primary');
const emailField = document.querySelector('#inputEmail');
const rememberMe = document.querySelector('#rememberMe');

init();

function init() {
  if ('serviceWorker' in navigator && 'PushManager' in window) {    
    navigator.serviceWorker.register('service-worker.js')
      .then(bootstrapUI)
      .catch(function(error) {
        alert('Service Worker Error', error);
      });
  } else {
    alert('Notifications are not supported');
    primaryButton.textContent = 'Notifications are not supported';
  }
}

function bootstrapUI(serviceWorker) {
  emailField.value = getCookie("email");
  emailField.addEventListener('focus', () => {
    console.log("[email] focus")
    updateBtn();
  });

  emailField.addEventListener('blur', () => {
    console.log("[email] blur")
    updateBtn();
  });

  primaryButton.addEventListener('click', () => {
    console.log('[button] clicked'); 
    primaryButton.disabled = true;
    if (rememberMe.checked) {
      setCookie('email', emailField.value, 365);
    }
    toggleSubscription(serviceWorker)
      .catch(err => {
        console.log(err);
        alert(err);
      });
  });

  return isSubscribed(serviceWorker).then(subscribed => {
    managePrimaryButton(subscribed);
  });
}

function managePrimaryButton(subscribed) {
  if (Notification.permission === 'denied') {
    primaryButton.textContent = 'Browser Notifications are denied.';
    primaryButton.disabled = true;
    return;
  }
  primaryButton.textContent = subscribed ? "Unsubscribe" : "Subscribe"
  primaryButton.disabled = false;
}

function toggleSubscription(serviceWorker) {
  return isSubscribed(serviceWorker).then(subscribed => {
    let promise = (subscribed ? unsubscribe(serviceWorker) : subscribe(serviceWorker));
    promise.then( () => {
      return managePrimaryButton(!subscribed);
    });
  });
}

function isSubscribed(serviceWorker) {
  return serviceWorker.pushManager.getSubscription()
    .then(subscription => subscription !== null);
}

function subscribe(serviceWorker) {
  const applicationServerKey = urlB64ToUint8Array(applicationServerPublicKey);
  return serviceWorker.pushManager.subscribe({
    userVisibleOnly: true,
    applicationServerKey: applicationServerKey
  })
  .then(subscription => {
    return registerToServer(emailField.value, subscription, 'POST');
  });
}

function unsubscribe(serviceWorker) {
  return serviceWorker.pushManager.getSubscription()
  .then(subscription => {
    if (subscription) {
      registerToServer(emailField.value, subscription, 'DELETE');
      return subscription.unsubscribe();
    }
  });
}

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
