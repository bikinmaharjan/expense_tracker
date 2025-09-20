import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import DocumentsView from '../views/DocumentsView.vue'
import { endpoints } from '../config/api'

describe('Documents upload', () => {
  it('submits form data with title and file', async () => {
    const axiosResponse = { data: {}, status: 201, statusText: 'Created', headers: {}, config: {} } as any
    const createMock = vi.spyOn(endpoints.documents, 'create').mockResolvedValue(axiosResponse)

    const wrapper = mount(DocumentsView, {
      global: { provide: { showNotification: () => {} } },
    })

    // Set form values directly
    (wrapper.vm as any).newDocument.title = 'Test Doc'
    (wrapper.vm as any).newDocument.description = 'desc'

    // Blob fallback in node; File may not be available in test env
    let fileObj: any
    try {
      fileObj = new File(['hello'], 'hello.txt', { type: 'text/plain' })
    } catch (e) {
      fileObj = new Blob(['hello'], { type: 'text/plain' })
    }
    (wrapper.vm as any).newDocument.file = fileObj

    await (wrapper.vm as any).handleSubmit()

    expect(createMock).toHaveBeenCalled()
    const formArg = (createMock as any).mock.calls[0][0]
    expect(formArg instanceof FormData).toBe(true)
    expect(formArg.get('title')).toBe('Test Doc')

    createMock.mockRestore()
  })
})
