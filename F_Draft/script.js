class DrawingApp {
    constructor() {
        this.canvas = document.getElementById('drawing-canvas');
        this.ctx = this.canvas.getContext('2d');
        this.isDrawing = false;
        this.lastX = 0;
        this.lastY = 0;
        
        // 工具设置
        this.currentTool = 'pencil';
        this.currentColor = '#000000';
        this.currentSize = 2; // 减小默认粗细
        
        // 画布尺寸设置
        this.originalWidth = 800;
        this.originalHeight = 600;
        this.isResizing = false;
        this.resizeStartX = 0;
        this.resizeStartY = 0;
        this.startWidth = 0;
        this.startHeight = 0;
        
        this.init();
    }
    
    init() {
        this.setupEventListeners();
        this.setupCanvas();
        this.updateToolButtons();
    }
    
    setupCanvas() {
        // 设置画布为白色背景
        this.ctx.fillStyle = 'white';
        this.ctx.fillRect(0, 0, this.canvas.width, this.canvas.height);
        this.ctx.lineJoin = 'round';
        this.ctx.lineCap = 'round';
    }
    
    setupEventListeners() {
        // 鼠标事件
        this.canvas.addEventListener('mousedown', this.startDrawing.bind(this));
        this.canvas.addEventListener('mousemove', this.draw.bind(this));
        this.canvas.addEventListener('mouseup', this.stopDrawing.bind(this));
        this.canvas.addEventListener('mouseout', this.stopDrawing.bind(this));
        
        // 触摸事件（移动设备支持）
        this.canvas.addEventListener('touchstart', this.handleTouchStart.bind(this));
        this.canvas.addEventListener('touchmove', this.handleTouchMove.bind(this));
        this.canvas.addEventListener('touchend', this.stopDrawing.bind(this));
        
        // 工具按钮事件
        document.querySelectorAll('.tool-btn[data-tool]').forEach(btn => {
            btn.addEventListener('click', (e) => {
                this.currentTool = e.target.dataset.tool;
                this.updateToolButtons();
            });
        });
        
        // 清除按钮
        document.getElementById('clear').addEventListener('click', () => {
            this.clearCanvas();
        });
        
        // 颜色选择器
        document.getElementById('color').addEventListener('input', (e) => {
            this.currentColor = e.target.value;
        });
        
        // 大小选择器
        document.getElementById('size').addEventListener('input', (e) => {
            this.currentSize = parseInt(e.target.value);
            document.getElementById('size-value').textContent = this.currentSize + 'px';
        });
        
        // 初始化默认粗细显示
        document.getElementById('size-value').textContent = this.currentSize + 'px';
        document.getElementById('size').value = this.currentSize;
        
        // 应用尺寸按钮
        document.getElementById('resize-canvas').addEventListener('click', () => {
            this.resizeCanvasByInput();
        });
        
        // 重置尺寸按钮
        document.getElementById('reset-canvas').addEventListener('click', () => {
            this.resetCanvasSize();
        });
        
        // 边框拉伸手柄
        const resizeHandle = document.querySelector('.resize-handle');
        resizeHandle.addEventListener('mousedown', this.startResize.bind(this));
        
        // 全局鼠标事件用于拉伸
        document.addEventListener('mousemove', this.handleResize.bind(this));
        document.addEventListener('mouseup', this.stopResize.bind(this));
    }
    
    startDrawing(e) {
        this.isDrawing = true;
        const pos = this.getMousePos(e);
        [this.lastX, this.lastY] = [pos.x, pos.y];
        
        // 如果是橡皮擦，开始擦除
        if (this.currentTool === 'eraser') {
            this.ctx.globalCompositeOperation = 'destination-out';
        } else {
            this.ctx.globalCompositeOperation = 'source-over';
        }
    }
    
    draw(e) {
        if (!this.isDrawing) return;
        
        e.preventDefault();
        const pos = this.getMousePos(e);
        
        this.ctx.beginPath();
        this.ctx.moveTo(this.lastX, this.lastY);
        this.ctx.lineTo(pos.x, pos.y);
        
        // 设置画笔属性
        if (this.currentTool === 'eraser') {
            this.ctx.strokeStyle = 'rgba(0,0,0,1)'; // 橡皮擦颜色
            this.ctx.globalCompositeOperation = 'destination-out';
        } else {
            this.ctx.strokeStyle = this.currentColor;
            this.ctx.globalCompositeOperation = 'source-over';
        }
        
        this.ctx.lineWidth = this.currentSize;
        this.ctx.stroke();
        
        [this.lastX, this.lastY] = [pos.x, pos.y];
    }
    
    stopDrawing() {
        this.isDrawing = false;
    }
    
    handleTouchStart(e) {
        e.preventDefault();
        const touch = e.touches[0];
        const mouseEvent = new MouseEvent('mousedown', {
            clientX: touch.clientX,
            clientY: touch.clientY
        });
        this.canvas.dispatchEvent(mouseEvent);
    }
    
    handleTouchMove(e) {
        e.preventDefault();
        const touch = e.touches[0];
        const mouseEvent = new MouseEvent('mousemove', {
            clientX: touch.clientX,
            clientY: touch.clientY
        });
        this.canvas.dispatchEvent(mouseEvent);
    }
    
    getMousePos(e) {
        const rect = this.canvas.getBoundingClientRect();
        let clientX, clientY;
        
        if (e.type.includes('touch')) {
            clientX = e.touches[0].clientX;
            clientY = e.touches[0].clientY;
        } else {
            clientX = e.clientX;
            clientY = e.clientY;
        }
        
        return {
            x: clientX - rect.left,
            y: clientY - rect.top
        };
    }
    
    updateToolButtons() {
        // 移除所有激活状态
        document.querySelectorAll('.tool-btn').forEach(btn => {
            btn.classList.remove('active');
        });
        
        // 激活当前工具
        document.querySelector(`[data-tool="${this.currentTool}"]`).classList.add('active');
    }
    
    clearCanvas() {
        if (confirm('确定要清除画布吗？')) {
            this.ctx.fillStyle = 'white';
            this.ctx.fillRect(0, 0, this.canvas.width, this.canvas.height);
        }
    }
    
    // 边框拉伸功能
    startResize(e) {
        e.preventDefault();
        this.isResizing = true;
        this.resizeStartX = e.clientX;
        this.resizeStartY = e.clientY;
        this.startWidth = this.canvas.width;
        this.startHeight = this.canvas.height;
        
        // 改变鼠标样式
        document.body.style.cursor = 'nw-resize';
        this.canvas.style.cursor = 'nw-resize';
    }
    
    handleResize(e) {
        if (!this.isResizing) return;
        
        const deltaX = e.clientX - this.resizeStartX;
        const deltaY = e.clientY - this.resizeStartY;
        
        // 计算新尺寸（最小100px）
        const newWidth = Math.max(100, this.startWidth + deltaX);
        const newHeight = Math.max(100, this.startHeight + deltaY);
        
        // 更新画布尺寸
        this.resizeCanvas(newWidth, newHeight);
    }
    
    stopResize() {
        if (!this.isResizing) return;
        
        this.isResizing = false;
        document.body.style.cursor = '';
        this.canvas.style.cursor = 'crosshair';
    }
    
    resizeCanvas(width, height) {
        // 保存当前画布内容
        const tempCanvas = document.createElement('canvas');
        tempCanvas.width = this.canvas.width;
        tempCanvas.height = this.canvas.height;
        const tempCtx = tempCanvas.getContext('2d');
        tempCtx.drawImage(this.canvas, 0, 0);
        
        // 设置新尺寸
        this.canvas.width = width;
        this.canvas.height = height;
        this.originalWidth = width;
        this.originalHeight = height;
        
        // 恢复画布内容
        this.ctx.fillStyle = 'white';
        this.ctx.fillRect(0, 0, width, height);
        this.ctx.drawImage(tempCanvas, 0, 0);
        
        // 更新输入框显示
        document.getElementById('canvas-width').value = width;
        document.getElementById('canvas-height').value = height;
    }
    
    resizeCanvasByInput() {
        const width = parseInt(document.getElementById('canvas-width').value) || 800;
        const height = parseInt(document.getElementById('canvas-height').value) || 600;
        
        // 限制尺寸范围
        const newWidth = Math.max(100, Math.min(2000, width));
        const newHeight = Math.max(100, Math.min(2000, height));
        
        this.resizeCanvas(newWidth, newHeight);
    }
    
    resetCanvasSize() {
        this.resizeCanvas(800, 600);
    }
    
    // 保存画布为图片
    saveCanvas() {
        const link = document.createElement('a');
        link.download = 'drawing-' + new Date().toISOString().slice(0, 19) + '.png';
        link.href = this.canvas.toDataURL();
        link.click();
    }
}

// 初始化应用
let app;
document.addEventListener('DOMContentLoaded', () => {
    app = new DrawingApp();
    
    // 添加快捷键支持
    document.addEventListener('keydown', (e) => {
        if (e.ctrlKey || e.metaKey) {
            switch(e.key) {
                case 's':
                    e.preventDefault();
                    app.saveCanvas();
                    break;
                case 'z':
                    e.preventDefault();
                    alert('撤销功能需要更复杂的实现');
                    break;
            }
        }
    });
});