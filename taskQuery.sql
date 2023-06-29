-- Return Spots which have a domain with count greater than 1
-- website field only contain the domain name
-- How many spots have the same domain
-- parse website string with regex to extract domain_name
SELECT 
    LOWER(name) AS spot_name, 
    LOWER(domain_name) domain_name, 
    COUNT(domain_name) AS spots_with_this_domain 
FROM (
	SELECT 
        name, 
		regexp_replace(
			regexp_replace(
				website, 
				'(?:http[s]?://)?(?:www.)?', '', 'gi'),
			'(/.*)', '', 'gi'
		) AS domain_name FROM "SPOTS") AS foo
	GROUP BY domain_name, spot_name
	HAVING COUNT(domain_name) > 1;

