package server

func (s *Server) testConnectionToHtmlcoind() error {
	_, err := s.htmlcoinRPCClient.GetNetworkInfo()
	return err
}
