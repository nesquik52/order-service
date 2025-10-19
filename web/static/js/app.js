class OrderService {
    constructor() {
        this.baseUrl = '';
        this.init();
    }

    init() {
        this.bindEvents();
        this.checkInitialOrder();
    }

    bindEvents() {
        const searchForm = document.getElementById('searchForm');
        const orderIdInput = document.getElementById('orderId');
        const testOrderBtn = document.getElementById('testOrderBtn');

        searchForm.addEventListener('submit', (e) => {
            e.preventDefault();
            this.searchOrder(orderIdInput.value.trim());
        });

        testOrderBtn.addEventListener('click', () => {
            orderIdInput.value = 'b563feb7b2b84b6test';
            this.searchOrder('b563feb7b2b84b6test');
        });

        orderIdInput.addEventListener('input', () => {
            this.hideResult();
        });
    }

    checkInitialOrder() {
        const urlParams = new URLSearchParams(window.location.search);
        const orderId = urlParams.get('id');
        if (orderId) {
            document.getElementById('orderId').value = orderId;
            this.searchOrder(orderId);
        }
    }

    async searchOrder(orderId) {
        if (!orderId) {
            this.showError('Please enter an Order ID');
            return;
        }

        this.showLoading();

        try {
            const response = await fetch(`/order?id=${encodeURIComponent(orderId)}`);
            
            if (!response.ok) {
                if (response.status === 404) {
                    throw new Error('Order not found');
                } else {
                    throw new Error('Server error');
                }
            }

            const order = await response.json();
            this.displayOrder(order);
            
        } catch (error) {
            this.showError(error.message);
        }
    }

    displayOrder(order) {
        const resultSection = document.getElementById('resultSection');
        resultSection.innerHTML = this.generateOrderHTML(order);
        resultSection.classList.add('active');
        
        resultSection.scrollIntoView({ behavior: 'smooth', block: 'start' });
    }

    generateOrderHTML(order) {
        return `
            <div class="order-header">
                <div class="order-id">Order #${order.order_uid}</div>
                <div class="order-status">${this.getStatusBadge(order)}</div>
            </div>
            
            <div class="order-content">
                <!-- Delivery Information -->
                <div class="section">
                    <div class="section-title">
                        Delivery Information
                    </div>
                    <div class="info-grid">
                        ${this.generateInfoGrid([
                            { label: 'Full Name', value: order.delivery.name },
                            { label: 'Phone', value: order.delivery.phone },
                            { label: 'Email', value: order.delivery.email },
                            { label: 'Address', value: `${order.delivery.city}, ${order.delivery.address}` },
                            { label: 'ZIP Code', value: order.delivery.zip },
                            { label: 'Region', value: order.delivery.region }
                        ])}
                    </div>
                </div>

                <!-- Payment Information -->
                <div class="section">
                    <div class="section-title">
                        Payment Information
                    </div>
                    <div class="info-grid">
                        ${this.generateInfoGrid([
                            { label: 'Transaction ID', value: order.payment.transaction },
                            { label: 'Amount', value: `$${order.payment.amount} ${order.payment.currency}` },
                            { label: 'Provider', value: order.payment.provider },
                            { label: 'Bank', value: order.payment.bank },
                            { label: 'Delivery Cost', value: `$${order.payment.delivery_cost}` },
                            { label: 'Goods Total', value: `$${order.payment.goods_total}` }
                        ])}
                    </div>
                </div>

                <!-- Items -->
                <div class="section">
                    <div class="section-title">
                        Order Items (${order.items.length})
                    </div>
                    <table class="items-table">
                        <thead>
                            <tr>
                                <th>Product</th>
                                <th>Brand</th>
                                <th>Price</th>
                                <th>Sale</th>
                                <th>Total</th>
                                <th>Status</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${order.items.map(item => `
                                <tr>
                                    <td><strong>${item.name}</strong></td>
                                    <td>${item.brand}</td>
                                    <td>$${item.price}</td>
                                    <td>${item.sale}%</td>
                                    <td class="amount">$${item.total_price}</td>
                                    <td><span class="status-badge status-delivered">${this.getItemStatus(item.status)}</span></td>
                                </tr>
                            `).join('')}
                        </tbody>
                    </table>
                </div>

                <!-- Order Details -->
                <div class="section">
                    <div class="section-title">
                        Order Details
                    </div>
                    <div class="info-grid">
                        ${this.generateInfoGrid([
                            { label: 'Track Number', value: order.track_number },
                            { label: 'Entry', value: order.entry },
                            { label: 'Customer ID', value: order.customer_id },
                            { label: 'Delivery Service', value: order.delivery_service },
                            { label: 'Locale', value: order.locale.toUpperCase() },
                            { label: 'Date Created', value: new Date(order.date_created).toLocaleString() }
                        ])}
                    </div>
                </div>
            </div>
        `;
    }

    generateInfoGrid(items) {
        return items.map(item => `
            <div class="info-item">
                <div class="info-label">${item.label}</div>
                <div class="info-value">${item.value || 'N/A'}</div>
            </div>
        `).join('');
    }

    getStatusBadge(order) {
        return '<span class="status-badge status-delivered">Completed</span>';
    }

    getItemStatus(statusCode) {
        const statusMap = {
            202: 'Completed',
            200: 'Processing',
            400: 'Cancelled'
        };
        return statusMap[statusCode] || 'Unknown';
    }

    showLoading() {
        const resultSection = document.getElementById('resultSection');
        resultSection.innerHTML = `
            <div class="loading">
                <div class="loading-spinner"></div>
                <div>Loading order information...</div>
            </div>
        `;
        resultSection.classList.add('active');
    }

    showError(message) {
        const resultSection = document.getElementById('resultSection');
        resultSection.innerHTML = `
            <div class="error-message">
                <strong>Error:</strong> ${message}
            </div>
        `;
        resultSection.classList.add('active');
    }

    hideResult() {
        const resultSection = document.getElementById('resultSection');
        resultSection.classList.remove('active');
    }
}

document.addEventListener('DOMContentLoaded', () => {
    new OrderService();
});