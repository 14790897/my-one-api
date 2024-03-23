import { API, showError } from '../helpers';

export async function getOAuthState() {
  const res = await API.get('/api/oauth/state');
  const { success, message, data } = res.data;
  if (success) {
    return data;
  } else {
    showError(message);
    return '';
  }
}

export async function onGitHubOAuthClicked(github_client_id) {
  const state = await getOAuthState();
  if (!state) return;
  location.href = `https://github.com/login/oauth/authorize?client_id=${github_client_id}&state=${state}&scope=user:email`;
}

export async function onLinuxDoOAuthClicked(linuxdo_client_id) {
  const state = await getOAuthState();
  if (!state) return;
  location.href = `https://connect.linux.do/oauth2/authorize?client_id=${linuxdo_client_id}&response_type=code&state=${state}&scope=user:profile`;
}
