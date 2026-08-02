-- Add steps/faq JSON columns to landing page and translation tables.
ALTER TABLE campaign_landing_pages
  ADD COLUMN steps JSON NULL COMMENT 'How to participate steps',
  ADD COLUMN faq JSON NULL COMMENT 'Frequently asked questions';

ALTER TABLE campaign_landing_page_translations
  ADD COLUMN steps JSON NULL COMMENT 'How to participate steps',
  ADD COLUMN faq JSON NULL COMMENT 'Frequently asked questions';
