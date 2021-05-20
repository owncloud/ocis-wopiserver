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
    },
    {
      extension: 'ods',
      routeName: 'wopi-edit',
      icon: 'x-office-document'
    },
    {
      extension: 'odp',
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
      app: App
    }
  }
]

const navItems = []

export default {
  appInfo,
  store,
  routes,
  navItems
}
