import 'regenerator-runtime/runtime'
import App from './components/App.vue'

import store from './store'

const appInfo = {
  name: 'Wopi',
  id: 'wopi',
  isFileEditor: true,
  icon: 'x-office-document',
  extensions: getExtensions(openModes, fileExtensions)
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


const fileExtensions = [
  'odt',
  'ott',
  'ods',
  'odp',
  'odg',
  'otg',
  'doc',
  'dot',
  'xls',
  'xlt',
  'xlm',
  'ppt',
  'pot',
  'pps',
  'vsd',
  'dxf',
  'wmf',
  'cdr',
  'pages',
  'number',
  'key'
]

const openModes = [
  'edit'
]

function getExtension (openMode, fileExtension) {
  return {
    extension: fileExtension,
    routeName: 'wopi-' + openMode,
    icon: 'x-office-document',
    newFileMenu: {
      menuTitle ($gettext) {
        return $gettext('New ' + fileExtension.toUpperCase() + ' document')
      }
    }
  }
}

function getExtensions (openModes, fileExtensions) {
  const ext = []
  openModes.forEach(
    function (m) {
      fileExtensions.forEach(
        function (e) {
          ext.push(getExtension(m, e)
          )
        }
      )
    }
  )
  return ext
}
