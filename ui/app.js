import 'regenerator-runtime/runtime'
import App from './components/App.vue'

import store from './store'

const appInfo = {
  name: 'Wopi',
  id: 'wopi',
  isFileEditor: true,
  icon: 'x-office-document',
  extensions: [
    {
      extension: 'odt',
      routeName: 'wopi-edit',
      icon: 'x-office-document'
    }
  ]
}

const routes = [
  {
    name: 'edit',
    path: '/edit/:filePath',
    components: {
      fullscreen: App
    }
  },
  {
    name: 'open',
    path: '/',
    components: {
      app: App
    }
  }
]

const navItems = [
  {
    name: 'Wopi',
    iconMaterial: appInfo.icon,
    route: {
      name: 'open',
      path: `/${appInfo.id}/`
    }
  }
]

export default {
  appInfo,
  store,
  routes,
  navItems
}
