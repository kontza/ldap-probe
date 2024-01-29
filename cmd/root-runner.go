package cmd

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"os"

	"github.com/go-ldap/ldap/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func rootRunner(cmd *cobra.Command, args []string) {
	if len(os.Args) < 2 {
		log.Fatal().Msg("Gimme a search term to work on!")
	}
	dialUrl := viper.GetString("dial-url")
	log.Info().Str("dial URL", dialUrl).Msg("Configured")
	l, err := ldap.DialURL(dialUrl)
	if err != nil {
		log.Fatal().Err(err).Msg("DialURL failed due to")
	}
	defer l.Close()

	// Read password from ~/.adpass
	adPass, err := os.ReadFile(os.ExpandEnv("$HOME/.adpass"))
	if err != nil {
		log.Fatal().Err(err).Msg("ReadFile failed due to")
	}

	bindDn := viper.GetString("bind-dn")
	log.Info().Str("bind DN", bindDn).Msg("Configured")
	err = l.Bind(bindDn, string(adPass))
	if err != nil {
		log.Fatal().Err(err).Msg("Bind failed due to")
	}

	baseDn := viper.GetString("base-dn")
	log.Info().Str("base DN", baseDn).Msg("Configured")
	subQuery := fmt.Sprintf("(&(objectclass=user)(|(sAMAccountName=%s)(userPrincipalName=%s)))", os.Args[1], os.Args[1])
	log.Info().Str("sub query", subQuery).Msg("Using")
	searchRequest := ldap.NewSearchRequest(
		baseDn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		subQuery,
		[]string{"dn", "cn", "phone"},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		log.Fatal().Err(err).Msg("Search failed due to")
	}

	for _, entry := range sr.Entries {
		log.Info().Msg("Result:")
		log.Info().Str("  DN", entry.DN).Send()
		log.Info().Interface("  CN", entry.GetAttributeValue("cn")).Send()
	}
}
