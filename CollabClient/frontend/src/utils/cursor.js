import { Extension } from '@tiptap/core'
import { Plugin, PluginKey } from '@tiptap/pm/state'
import { Decoration, DecorationSet } from '@tiptap/pm/view'

export const cursorPluginKey = new PluginKey('remote-cursor')

export const RemoteCursor = Extension.create({
    name: 'remoteCursor',

    addProseMirrorPlugins() {
        return [
            new Plugin({
                key: cursorPluginKey,
                state: {
                    // 初始化：空数据
                    init() {
                        return { cursors: [] }
                    },
                    // 事务处理：只处理数据更新
                    apply(tr, prev) {
                        const meta = tr.getMeta(cursorPluginKey)

                        // 1. 如果有新数据传来，直接替换
                        if (meta && meta.type === 'update' && Array.isArray(meta.cursors)) {
                            return { cursors: meta.cursors }
                        }

                        // 2. 如果没有新数据，需要把旧光标位置映射到新文档位置
                        // (例如：我在前面打字，别人的光标应该往后退)
                        return {
                            cursors: prev.cursors.map(c => ({
                                ...c,
                                pos: tr.mapping.map(c.pos)
                            }))
                        }
                    }
                },
                props: {
                    // 渲染逻辑：根据 State 数据动态构建装饰器
                    decorations(state) {
                        const { cursors } = this.getState(state)
                        const docSize = state.doc.content.size
                        const decorations = []

                        for (const cursor of cursors) {
                            // 安全检查
                            if (cursor.pos === null || cursor.pos === undefined) continue
                            // 越界检查 (防止 crash 的最后一道防线)
                            let pos = cursor.pos
                            if (pos < 0) pos = 0
                            if (pos > docSize) pos = docSize

                            // 创建 DOM
                            const cursorElement = document.createElement('span')
                            cursorElement.classList.add('remote-cursor')
                            cursorElement.style.borderLeftColor = cursor.color

                            const labelElement = document.createElement('div')
                            labelElement.classList.add('remote-cursor-label')
                            labelElement.style.backgroundColor = cursor.color
                            labelElement.textContent = cursor.name
                            cursorElement.appendChild(labelElement)

                            decorations.push(Decoration.widget(pos, cursorElement, {
                                key: cursor.id,
                                side: -1
                            }))
                        }

                        return DecorationSet.create(state.doc, decorations)
                    }
                }
            })
        ]
    }
})