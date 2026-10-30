<?php

declare(strict_types=1);

class CurlHelper
{
    private string $baseUrl;
    private int $timeout;
    private ?string $token;

    /**
     * @param string|null $token JWT da API (ex.: $_SESSION['token']) para rotas protegidas
     */
    public function __construct(?string $token = null)
    {
        $this->baseUrl = rtrim(getenv('API_BASE_URL') ?: 'http://nginx/api', '/');
        $this->timeout = (int) (getenv('API_TIMEOUT') ?: 10);
        $this->token   = $token;
    }

    public function post(string $endpoint, array $data): array
    {
        return $this->request('POST', $endpoint, $data);
    }

    public function get(string $endpoint, array $params = []): array
    {
        $url = $this->baseUrl . $endpoint;
        if (!empty($params)) {
            $url .= '?' . http_build_query($params);
        }
        return $this->request('GET', $url);
    }

    private function request(string $method, string $endpoint, array $data = []): array
    {
        $url = str_starts_with($endpoint, 'http') ? $endpoint : $this->baseUrl . $endpoint;
        $ch  = curl_init($url);

        $headers = [
            'Content-Type: application/json',
            'Accept: application/json',
        ];
        if ($this->token !== null && $this->token !== '') {
            $headers[] = 'Authorization: Bearer ' . $this->token;
        }

        $options = [
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_TIMEOUT        => $this->timeout,
            CURLOPT_HTTPHEADER     => $headers,
        ];

        if ($method === 'POST') {
            $options[CURLOPT_POST]       = true;
            $options[CURLOPT_POSTFIELDS] = json_encode($data);
        }

        curl_setopt_array($ch, $options);

        $body     = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        $curlErr  = curl_error($ch);
        curl_close($ch);

        if ($body === false) {
            return ['body' => null, 'status' => 0, 'error' => $curlErr];
        }

        return ['body' => json_decode($body, true), 'status' => $httpCode];
    }
}
