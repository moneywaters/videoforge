package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/videoforge/backend/svc-shopify/internal/model"
)

// ShopifyRepository defines the interface for Shopify data access
type ShopifyRepository interface {
	// Store operations
	CreateStore(ctx context.Context, store *model.ShopifyStore) error
	GetStoreByID(ctx context.Context, id string) (*model.ShopifyStore, error)
	GetStoreByDomain(ctx context.Context, domain string) (*model.ShopifyStore, error)
	GetStoresByClient(ctx context.Context, clientID string) ([]model.ShopifyStore, error)
	UpdateStoreStatus(ctx context.Context, id, status string) error

	// Link operations
	CreateLink(ctx context.Context, link *model.VideoLink) error
	GetLinkByID(ctx context.Context, id string) (*model.VideoLink, error)
	GetLinkByDiscountCode(ctx context.Context, code string) (*model.VideoLink, error)
	GetLinksByVideo(ctx context.Context, videoID string) ([]model.VideoLink, error)
	GetLinksByCampaign(ctx context.Context, campaignID string) ([]model.VideoLink, error)
	GetLinksByUTMSource(ctx context.Context, source string) ([]model.VideoLink, error)
	GetLinksByUTMCampaign(ctx context.Context, campaign string) ([]model.VideoLink, error)
	GetAllLinks(ctx context.Context) ([]model.VideoLink, error)

	// Order operations
	CreateOrder(ctx context.Context, order *model.Order) error
	GetOrderByID(ctx context.Context, id string) (*model.Order, error)
	GetOrderByShopifyID(ctx context.Context, shopifyID string, storeID string) (*model.Order, error)
	GetOrdersByStore(ctx context.Context, storeID string) ([]model.Order, error)
	GetOrdersByVideo(ctx context.Context, videoID string) ([]model.Order, error)
	GetOrdersByStatus(ctx context.Context, status string) ([]model.Order, error)
	GetAllOrders(ctx context.Context) ([]model.Order, error)
	UpdateOrderStatus(ctx context.Context, id, status string) error

	// Attribution operations
	CreateAttribution(ctx context.Context, attr *model.Attribution) error
	GetAttributionByOrder(ctx context.Context, orderID string) (*model.Attribution, error)
	GetAttributionsByVideo(ctx context.Context, videoID string) ([]model.Attribution, error)
	GetAttributionsByCampaign(ctx context.Context, campaignID string) ([]model.Attribution, error)
	GetAllAttributions(ctx context.Context) ([]model.Attribution, error)
	GetAttributionSummaryByVideo(ctx context.Context, videoID string) (*model.AttributionSummary, error)
	GetAttributionSummaryByCampaign(ctx context.Context, campaignID string) (*model.AttributionSummary, error)
	GetAllAttributionSummaries(ctx context.Context) ([]model.AttributionSummary, error)
}

// PgShopifyRepository implements ShopifyRepository using pgx
type PgShopifyRepository struct {
	pool *pgxpool.Pool
}

// NewShopifyRepository creates a new Shopify repository
func NewShopifyRepository(pool *pgxpool.Pool) *PgShopifyRepository {
	return &PgShopifyRepository{pool: pool}
}

// =============================================================================
// Store Operations
// =============================================================================

func (r *PgShopifyRepository) CreateStore(ctx context.Context, store *model.ShopifyStore) error {
	if store.ID == "" {
		store.ID = uuid.New().String()
	}
	query := `
		INSERT INTO shopify_stores (id, client_id, shop_domain, access_token, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (shop_domain) DO UPDATE SET
			access_token = EXCLUDED.access_token,
			status = EXCLUDED.status,
			updated_at = NOW()
	`
	_, err := r.pool.Exec(ctx, query,
		store.ID,
		store.ClientID,
		store.ShopDomain,
		store.AccessToken,
		store.Status,
	)
	return err
}

func (r *PgShopifyRepository) GetStoreByID(ctx context.Context, id string) (*model.ShopifyStore, error) {
	query := `
		SELECT id, client_id, shop_domain, access_token, status, created_at, updated_at
		FROM shopify_stores
		WHERE id = $1
	`
	var store model.ShopifyStore
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&store.ID,
		&store.ClientID,
		&store.ShopDomain,
		&store.AccessToken,
		&store.Status,
		&store.CreatedAt,
		&store.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &store, nil
}

func (r *PgShopifyRepository) GetStoreByDomain(ctx context.Context, domain string) (*model.ShopifyStore, error) {
	query := `
		SELECT id, client_id, shop_domain, access_token, status, created_at, updated_at
		FROM shopify_stores
		WHERE shop_domain = $1
	`
	var store model.ShopifyStore
	err := r.pool.QueryRow(ctx, query, domain).Scan(
		&store.ID,
		&store.ClientID,
		&store.ShopDomain,
		&store.AccessToken,
		&store.Status,
		&store.CreatedAt,
		&store.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &store, nil
}

func (r *PgShopifyRepository) GetStoresByClient(ctx context.Context, clientID string) ([]model.ShopifyStore, error) {
	query := `
		SELECT id, client_id, shop_domain, access_token, status, created_at, updated_at
		FROM shopify_stores
		WHERE client_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stores []model.ShopifyStore
	for rows.Next() {
		var store model.ShopifyStore
		if err := rows.Scan(
			&store.ID,
			&store.ClientID,
			&store.ShopDomain,
			&store.AccessToken,
			&store.Status,
			&store.CreatedAt,
			&store.UpdatedAt,
		); err != nil {
			return nil, err
		}
		stores = append(stores, store)
	}
	return stores, nil
}

func (r *PgShopifyRepository) UpdateStoreStatus(ctx context.Context, id, status string) error {
	query := `
		UPDATE shopify_stores
		SET status = $2, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, id, status)
	return err
}

// =============================================================================
// Link Operations
// =============================================================================

func (r *PgShopifyRepository) CreateLink(ctx context.Context, link *model.VideoLink) error {
	if link.ID == "" {
		link.ID = uuid.New().String()
	}
	query := `
		INSERT INTO video_links (id, video_id, campaign_id, discount_code, utm_source, utm_medium, utm_campaign, url, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		ON CONFLICT (discount_code) DO UPDATE SET
			video_id = EXCLUDED.video_id,
			campaign_id = EXCLUDED.campaign_id,
			utm_source = EXCLUDED.utm_source,
			utm_medium = EXCLUDED.utm_medium,
			utm_campaign = EXCLUDED.utm_campaign,
			url = EXCLUDED.url
	`
	_, err := r.pool.Exec(ctx, query,
		link.ID,
		link.VideoID,
		link.CampaignID,
		link.DiscountCode,
		link.UTMSource,
		link.UTMMedium,
		link.UTMCampaign,
		link.URL,
	)
	return err
}

func (r *PgShopifyRepository) GetLinkByID(ctx context.Context, id string) (*model.VideoLink, error) {
	query := `
		SELECT id, video_id, campaign_id, discount_code, utm_source, utm_medium, utm_campaign, url, created_at
		FROM video_links
		WHERE id = $1
	`
	var link model.VideoLink
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&link.ID,
		&link.VideoID,
		&link.CampaignID,
		&link.DiscountCode,
		&link.UTMSource,
		&link.UTMMedium,
		&link.UTMCampaign,
		&link.URL,
		&link.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *PgShopifyRepository) GetLinkByDiscountCode(ctx context.Context, code string) (*model.VideoLink, error) {
	query := `
		SELECT id, video_id, campaign_id, discount_code, utm_source, utm_medium, utm_campaign, url, created_at
		FROM video_links
		WHERE discount_code = $1
	`
	var link model.VideoLink
	err := r.pool.QueryRow(ctx, query, code).Scan(
		&link.ID,
		&link.VideoID,
		&link.CampaignID,
		&link.DiscountCode,
		&link.UTMSource,
		&link.UTMMedium,
		&link.UTMCampaign,
		&link.URL,
		&link.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *PgShopifyRepository) GetLinksByVideo(ctx context.Context, videoID string) ([]model.VideoLink, error) {
	query := `
		SELECT id, video_id, campaign_id, discount_code, utm_source, utm_medium, utm_campaign, url, created_at
		FROM video_links
		WHERE video_id = $1
		ORDER BY created_at DESC
	`
	return r.queryLinks(ctx, query, videoID)
}

func (r *PgShopifyRepository) GetLinksByCampaign(ctx context.Context, campaignID string) ([]model.VideoLink, error) {
	query := `
		SELECT id, video_id, campaign_id, discount_code, utm_source, utm_medium, utm_campaign, url, created_at
		FROM video_links
		WHERE campaign_id = $1
		ORDER BY created_at DESC
	`
	return r.queryLinks(ctx, query, campaignID)
}

func (r *PgShopifyRepository) GetLinksByUTMSource(ctx context.Context, source string) ([]model.VideoLink, error) {
	query := `
		SELECT id, video_id, campaign_id, discount_code, utm_source, utm_medium, utm_campaign, url, created_at
		FROM video_links
		WHERE utm_source = $1
		ORDER BY created_at DESC
	`
	return r.queryLinks(ctx, query, source)
}

func (r *PgShopifyRepository) GetLinksByUTMCampaign(ctx context.Context, campaign string) ([]model.VideoLink, error) {
	query := `
		SELECT id, video_id, campaign_id, discount_code, utm_source, utm_medium, utm_campaign, url, created_at
		FROM video_links
		WHERE utm_campaign = $1
		ORDER BY created_at DESC
	`
	return r.queryLinks(ctx, query, campaign)
}

func (r *PgShopifyRepository) GetAllLinks(ctx context.Context) ([]model.VideoLink, error) {
	query := `
		SELECT id, video_id, campaign_id, discount_code, utm_source, utm_medium, utm_campaign, url, created_at
		FROM video_links
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []model.VideoLink
	for rows.Next() {
		var link model.VideoLink
		if err := rows.Scan(
			&link.ID,
			&link.VideoID,
			&link.CampaignID,
			&link.DiscountCode,
			&link.UTMSource,
			&link.UTMMedium,
			&link.UTMCampaign,
			&link.URL,
			&link.CreatedAt,
		); err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	return links, nil
}

func (r *PgShopifyRepository) queryLinks(ctx context.Context, query string, arg interface{}) ([]model.VideoLink, error) {
	rows, err := r.pool.Query(ctx, query, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []model.VideoLink
	for rows.Next() {
		var link model.VideoLink
		if err := rows.Scan(
			&link.ID,
			&link.VideoID,
			&link.CampaignID,
			&link.DiscountCode,
			&link.UTMSource,
			&link.UTMMedium,
			&link.UTMCampaign,
			&link.URL,
			&link.CreatedAt,
		); err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	return links, nil
}

// =============================================================================
// Order Operations
// =============================================================================

func (r *PgShopifyRepository) CreateOrder(ctx context.Context, order *model.Order) error {
	if order.ID == "" {
		order.ID = uuid.New().String()
	}
	query := `
		INSERT INTO orders (id, shopify_order_id, store_id, customer_email, total_price, currency, discount_code, utm_source, utm_medium, utm_campaign, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
		ON CONFLICT (shopify_order_id, store_id) DO UPDATE SET
			customer_email = EXCLUDED.customer_email,
			total_price = EXCLUDED.total_price,
			discount_code = EXCLUDED.discount_code,
			utm_source = EXCLUDED.utm_source,
			utm_medium = EXCLUDED.utm_medium,
			utm_campaign = EXCLUDED.utm_campaign,
			status = EXCLUDED.status,
			updated_at = NOW()
	`
	_, err := r.pool.Exec(ctx, query,
		order.ID,
		order.ShopifyOrderID,
		order.StoreID,
		order.CustomerEmail,
		order.TotalPrice,
		order.Currency,
		order.DiscountCode,
		order.UTMSource,
		order.UTMMedium,
		order.UTMCampaign,
		order.Status,
	)
	return err
}

func (r *PgShopifyRepository) GetOrderByID(ctx context.Context, id string) (*model.Order, error) {
	query := `
		SELECT id, shopify_order_id, store_id, customer_email, total_price, currency, discount_code, utm_source, utm_medium, utm_campaign, status, created_at, updated_at
		FROM orders
		WHERE id = $1
	`
	var order model.Order
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&order.ID,
		&order.ShopifyOrderID,
		&order.StoreID,
		&order.CustomerEmail,
		&order.TotalPrice,
		&order.Currency,
		&order.DiscountCode,
		&order.UTMSource,
		&order.UTMMedium,
		&order.UTMCampaign,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *PgShopifyRepository) GetOrderByShopifyID(ctx context.Context, shopifyID string, storeID string) (*model.Order, error) {
	query := `
		SELECT id, shopify_order_id, store_id, customer_email, total_price, currency, discount_code, utm_source, utm_medium, utm_campaign, status, created_at, updated_at
		FROM orders
		WHERE shopify_order_id = $1 AND store_id = $2
	`
	var order model.Order
	err := r.pool.QueryRow(ctx, query, shopifyID, storeID).Scan(
		&order.ID,
		&order.ShopifyOrderID,
		&order.StoreID,
		&order.CustomerEmail,
		&order.TotalPrice,
		&order.Currency,
		&order.DiscountCode,
		&order.UTMSource,
		&order.UTMMedium,
		&order.UTMCampaign,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *PgShopifyRepository) GetOrdersByStore(ctx context.Context, storeID string) ([]model.Order, error) {
	query := `
		SELECT id, shopify_order_id, store_id, customer_email, total_price, currency, discount_code, utm_source, utm_medium, utm_campaign, status, created_at, updated_at
		FROM orders
		WHERE store_id = $1
		ORDER BY created_at DESC
	`
	return r.queryOrders(ctx, query, storeID)
}

func (r *PgShopifyRepository) GetOrdersByVideo(ctx context.Context, videoID string) ([]model.Order, error) {
	query := `
		SELECT o.id, o.shopify_order_id, o.store_id, o.customer_email, o.total_price, o.currency, o.discount_code, o.utm_source, o.utm_medium, o.utm_campaign, o.status, o.created_at, o.updated_at
		FROM orders o
		INNER JOIN video_links vl ON vl.video_id = $1
		WHERE o.discount_code = ANY(ARRAY[vl.discount_code])
		ORDER BY o.created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, videoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var order model.Order
		if err := rows.Scan(
			&order.ID,
			&order.ShopifyOrderID,
			&order.StoreID,
			&order.CustomerEmail,
			&order.TotalPrice,
			&order.Currency,
			&order.DiscountCode,
			&order.UTMSource,
			&order.UTMMedium,
			&order.UTMCampaign,
			&order.Status,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (r *PgShopifyRepository) GetOrdersByStatus(ctx context.Context, status string) ([]model.Order, error) {
	query := `
		SELECT id, shopify_order_id, store_id, customer_email, total_price, currency, discount_code, utm_source, utm_medium, utm_campaign, status, created_at, updated_at
		FROM orders
		WHERE status = $1
		ORDER BY created_at DESC
	`
	return r.queryOrders(ctx, query, status)
}

func (r *PgShopifyRepository) GetAllOrders(ctx context.Context) ([]model.Order, error) {
	query := `
		SELECT id, shopify_order_id, store_id, customer_email, total_price, currency, discount_code, utm_source, utm_medium, utm_campaign, status, created_at, updated_at
		FROM orders
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var order model.Order
		if err := rows.Scan(
			&order.ID,
			&order.ShopifyOrderID,
			&order.StoreID,
			&order.CustomerEmail,
			&order.TotalPrice,
			&order.Currency,
			&order.DiscountCode,
			&order.UTMSource,
			&order.UTMMedium,
			&order.UTMCampaign,
			&order.Status,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (r *PgShopifyRepository) UpdateOrderStatus(ctx context.Context, id, status string) error {
	query := `
		UPDATE orders
		SET status = $2, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query, id, status)
	return err
}

func (r *PgShopifyRepository) queryOrders(ctx context.Context, query string, arg interface{}) ([]model.Order, error) {
	rows, err := r.pool.Query(ctx, query, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var order model.Order
		if err := rows.Scan(
			&order.ID,
			&order.ShopifyOrderID,
			&order.StoreID,
			&order.CustomerEmail,
			&order.TotalPrice,
			&order.Currency,
			&order.DiscountCode,
			&order.UTMSource,
			&order.UTMMedium,
			&order.UTMCampaign,
			&order.Status,
			&order.CreatedAt,
			&order.UpdatedAt,
		); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

// =============================================================================
// Attribution Operations
// =============================================================================

func (r *PgShopifyRepository) CreateAttribution(ctx context.Context, attr *model.Attribution) error {
	if attr.ID == "" {
		attr.ID = uuid.New().String()
	}
	query := `
		INSERT INTO attributions (id, order_id, video_id, campaign_id, attributed_amount, attribution_method, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`
	_, err := r.pool.Exec(ctx, query,
		attr.ID,
		attr.OrderID,
		attr.VideoID,
		attr.CampaignID,
		attr.AttributedAmount,
		attr.AttributionMethod,
	)
	return err
}

func (r *PgShopifyRepository) GetAttributionByOrder(ctx context.Context, orderID string) (*model.Attribution, error) {
	query := `
		SELECT id, order_id, video_id, campaign_id, attributed_amount, attribution_method, created_at
		FROM attributions
		WHERE order_id = $1
	`
	var attr model.Attribution
	err := r.pool.QueryRow(ctx, query, orderID).Scan(
		&attr.ID,
		&attr.OrderID,
		&attr.VideoID,
		&attr.CampaignID,
		&attr.AttributedAmount,
		&attr.AttributionMethod,
		&attr.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &attr, nil
}

func (r *PgShopifyRepository) GetAttributionsByVideo(ctx context.Context, videoID string) ([]model.Attribution, error) {
	query := `
		SELECT id, order_id, video_id, campaign_id, attributed_amount, attribution_method, created_at
		FROM attributions
		WHERE video_id = $1
		ORDER BY created_at DESC
	`
	return r.queryAttributions(ctx, query, videoID)
}

func (r *PgShopifyRepository) GetAttributionsByCampaign(ctx context.Context, campaignID string) ([]model.Attribution, error) {
	query := `
		SELECT id, order_id, video_id, campaign_id, attributed_amount, attribution_method, created_at
		FROM attributions
		WHERE campaign_id = $1
		ORDER BY created_at DESC
	`
	return r.queryAttributions(ctx, query, campaignID)
}

func (r *PgShopifyRepository) GetAllAttributions(ctx context.Context) ([]model.Attribution, error) {
	query := `
		SELECT id, order_id, video_id, campaign_id, attributed_amount, attribution_method, created_at
		FROM attributions
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attributions []model.Attribution
	for rows.Next() {
		var attr model.Attribution
		if err := rows.Scan(
			&attr.ID,
			&attr.OrderID,
			&attr.VideoID,
			&attr.CampaignID,
			&attr.AttributedAmount,
			&attr.AttributionMethod,
			&attr.CreatedAt,
		); err != nil {
			return nil, err
		}
		attributions = append(attributions, attr)
	}
	return attributions, nil
}

func (r *PgShopifyRepository) GetAttributionSummaryByVideo(ctx context.Context, videoID string) (*model.AttributionSummary, error) {
	query := `
		SELECT
			video_id,
			COALESCE(campaign_id::text, ''),
			SUM(total_price)::numeric(12,2) as total_sales,
			COUNT(*) as total_orders,
			SUM(attributed_amount)::numeric(12,2) as attributed_amount
		FROM attributions a
		INNER JOIN orders o ON o.id = a.order_id
		WHERE a.video_id = $1
		GROUP BY video_id, campaign_id
	`
	var summary model.AttributionSummary
	err := r.pool.QueryRow(ctx, query, videoID).Scan(
		&summary.VideoID,
		&summary.CampaignID,
		&summary.TotalSales,
		&summary.TotalOrders,
		&summary.AttributedAmount,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

func (r *PgShopifyRepository) GetAttributionSummaryByCampaign(ctx context.Context, campaignID string) (*model.AttributionSummary, error) {
	query := `
		SELECT
			video_id,
			COALESCE(campaign_id::text, ''),
			SUM(total_price)::numeric(12,2) as total_sales,
			COUNT(*) as total_orders,
			SUM(attributed_amount)::numeric(12,2) as attributed_amount
		FROM attributions a
		INNER JOIN orders o ON o.id = a.order_id
		WHERE a.campaign_id = $1
		GROUP BY video_id, campaign_id
	`
	var summary model.AttributionSummary
	err := r.pool.QueryRow(ctx, query, campaignID).Scan(
		&summary.VideoID,
		&summary.CampaignID,
		&summary.TotalSales,
		&summary.TotalOrders,
		&summary.AttributedAmount,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

func (r *PgShopifyRepository) GetAllAttributionSummaries(ctx context.Context) ([]model.AttributionSummary, error) {
	query := `
		SELECT
			a.video_id,
			COALESCE(a.campaign_id::text, ''),
			SUM(o.total_price)::numeric(12,2) as total_sales,
			COUNT(*) as total_orders,
			SUM(a.attributed_amount)::numeric(12,2) as attributed_amount
		FROM attributions a
		INNER JOIN orders o ON o.id = a.order_id
		GROUP BY a.video_id, a.campaign_id
		ORDER BY attributed_amount DESC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []model.AttributionSummary
	for rows.Next() {
		var summary model.AttributionSummary
		if err := rows.Scan(
			&summary.VideoID,
			&summary.CampaignID,
			&summary.TotalSales,
			&summary.TotalOrders,
			&summary.AttributedAmount,
		); err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}
	return summaries, nil
}

func (r *PgShopifyRepository) queryAttributions(ctx context.Context, query string, arg interface{}) ([]model.Attribution, error) {
	rows, err := r.pool.Query(ctx, query, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attributions []model.Attribution
	for rows.Next() {
		var attr model.Attribution
		if err := rows.Scan(
			&attr.ID,
			&attr.OrderID,
			&attr.VideoID,
			&attr.CampaignID,
			&attr.AttributedAmount,
			&attr.AttributionMethod,
			&attr.CreatedAt,
		); err != nil {
			return nil, err
		}
		attributions = append(attributions, attr)
	}
	return attributions, nil
}

// Ensure PgShopifyRepository implements ShopifyRepository
var _ ShopifyRepository = (*PgShopifyRepository)(nil)

// Helper for error wrapping
func wrapError(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}