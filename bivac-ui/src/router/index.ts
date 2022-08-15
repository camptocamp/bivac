import { createRouter, createWebHashHistory } from 'vue-router'
import InfoView from '../views/InfoView.vue'
import VolumeView from '../views/VolumeView.vue'
import VolumeDetailsView from '@/views/VolumeDetailsView.vue'

const router = createRouter({
  history: createWebHashHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      component: InfoView
    },
    {
      path: '/volumes',
      component: VolumeView,
    },
    {
      path: '/volume/:id',
      component: VolumeDetailsView
    }
  ]
})

export default router
