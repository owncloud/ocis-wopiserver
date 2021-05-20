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
      extension: 'ott',
      routeName: 'wopi-edit',
      icon: 'x-office-document'
    },
    {
      extension: 'ods',
      routeName: 'wopi-edit',
      icon: 'x-office-document'
    },
    {
      extension: 'ots',
      routeName: 'wopi-edit',
      icon: 'x-office-document'
    },
    {
      extension: 'odp',
      routeName: 'wopi-edit',
      icon: 'x-office-document'
    },
    {
      extension: 'otp',
      routeName: 'wopi-edit',
      icon: 'x-office-document'
    },
    {
      extension: 'odg',
      routeName: 'wopi-edit',
      icon: 'x-office-document'
    },
    {
      extension: 'otg',
      routeName: 'wopi-edit',
      icon: 'x-office-document'
    },
    {
      extension: 'doc',
      routeName: 'wopi-edit',
      icon: 'x-office-document'
    },
    {
      extension: 'dot',
      routeName: 'wopi-edit',
      icon: 'x-office-document'
    },
    {
      extension: 'xls',
      routeName: 'wopi-edit',
      icon: 'x-office-document'
    },
    {
      extension: 'xlt',
      routeName: 'wopi-edit',
      icon: 'x-office-document'
    },
    {
      extension: 'xlm',
      routeName: 'wopi-edit',
      icon: 'x-office-document'
    },
    {
      extension: 'ppt',
      routeName: 'wopi-edit',
      icon: 'x-office-document'
    },
    {
      extension: 'pot',
      routeName: 'wopi-edit',
      icon: 'x-office-document'
    },
    {
      extension: 'pps',
      routeName: 'wopi-edit',
      icon: 'x-office-document'
    },
    {
      extension: 'vsd',
      routeName: 'wopi-edit',
      icon: 'x-office-document'
    },
    {
      extension: 'dxf',
      routeName: 'wopi-edit',
      icon: 'x-office-document'
    },
    {
      extension: 'wmf',
      routeName: 'wopi-edit',
      icon: 'x-office-document'
    },
    {
      extension: 'cdr',
      routeName: 'wopi-edit',
      icon: 'x-office-document'
    },
    {
      extension: 'pages',
      routeName: 'wopi-edit',
      icon: 'x-office-document'
    },
    {
      extension: 'number',
      routeName: 'wopi-edit',
      icon: 'x-office-document'
    },
    {
      extension: 'key',
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
